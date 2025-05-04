package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/raft"
)

// RaftOrderService wraps OrderService to provide distributed consensus
type RaftOrderService struct {
	orderService  *OrderService
	raftNode      *raft.RaftNode
	applyCh       chan raft.LogEntry
	nodeID        string
	isLeader      bool
	logger        *log.Logger
	resultMap     map[uint64]*domain.Order
	resultMapLock sync.Mutex
}

// NewRaftOrderService creates a new Raft-enabled order service
func NewRaftOrderService(
	orderService *OrderService,
	nodeID string,
	peerIDs []string,
	peerAddrs map[string]string,
) (*RaftOrderService, error) {
	logger := log.New(os.Stdout, fmt.Sprintf("[RAFT-%s] ", nodeID), log.LstdFlags)

	applyCh := make(chan raft.LogEntry, raft.MaxLogEntriesBuffer)

	service := &RaftOrderService{
		orderService: orderService,
		applyCh:      applyCh,
		nodeID:       nodeID,
		isLeader:     false,
		logger:       logger,
		resultMap:    make(map[uint64]*domain.Order),
	}

	// Create the Raft node
	raftNode := raft.NewRaftNode(
		nodeID,
		peerIDs,
		peerAddrs,
		applyCh,
		// service.applyCommand,
		func(cmd interface{}) error {
			_, err := service.applyCommand(cmd)
			return err
		},
		logger,
	)

	service.raftNode = raftNode

	// Start processing applied commands
	go service.processAppliedCommands()

	return service, nil
}

// Start initializes and starts the Raft node
func (s *RaftOrderService) Start(ctx context.Context) error {
	// Start the cleanup goroutine
	go s.cleanupResults()

	// Start the Raft node
	return s.raftNode.Start(ctx)
}

// CreateOrder creates a new order with Raft consensus
func (s *RaftOrderService) CreateOrder(
	ctx context.Context,
	customerID, merchantID uint,
	items []SimpleItem,
	notes string,
) (*domain.Order, error) {
	// Prepare the order command
	// Convert the map slice to the expected type
	raftItems := make([]raft.OrderItemCommand, len(items))
	for i, item := range items {
		raftItems[i] = raft.OrderItemCommand{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	cmd := raft.OrderCommand{
		Type:       "create_order",
		CustomerID: customerID,
		MerchantID: merchantID,
		OrderItems: raftItems,
		AdditionalData: map[string]interface{}{
			"notes": notes,
		},
	}
	fmt.Printf("DEBUG: Submitting order to Raft: %v\n", cmd)
	// Submit the command to Raft
	index, err := s.raftNode.Submit(cmd)
	fmt.Printf("DEBUG: Submitted order to Raft: index=%d, err=%v\n", index, err)
	if err != nil {
		return nil, fmt.Errorf("failed to submit order to Raft: %w", err)
	}

	// Wait for the command to be applied
	fmt.Printf("DEBUG: Waiting for order to be applied\n")
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("DEBUG: Ticker ticked\n")
			// Check if the command has been applied
			s.updateLastApplied(index)

			// Check if the command has been applied
			s.resultMapLock.Lock()
			order, exists := s.resultMap[index]
			if exists {
				// Delete the result from the cache to avoid memory leaks
				delete(s.resultMap, index)
				s.resultMapLock.Unlock()
				fmt.Printf("DEBUG: Retrieved order from result cache: %v\n", order)
				return order, nil
			}
			s.resultMapLock.Unlock()

		case <-timeout:
			fmt.Printf("DEBUG: Timeout waiting for order creation\n")
			return nil, errors.New("timeout waiting for order creation")

		case <-ctx.Done():
			fmt.Printf("DEBUG: Context done\n")
			return nil, ctx.Err()
		}
	}
}

// applyCommand applies a Raft command to the state machine
func (s *RaftOrderService) applyCommand(cmdInterface interface{}) (*domain.Order, error) {
	var createdOrder *domain.Order = nil
	// Convert the interface to an OrderCommand
	cmdBytes, err := json.Marshal(cmdInterface)
	fmt.Printf("DEBUG: cmdBytes: %v\n", string(cmdBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	var cmd raft.OrderCommand
	if err := json.Unmarshal(cmdBytes, &cmd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}

	if s.nodeID != s.raftNode.LeaderID() {
		s.logger.Printf("[Follower-%s] skip %s (already done by leader %s)",
			s.nodeID, cmd.Type, s.raftNode.LeaderID())
		return nil, nil
	}

	fmt.Printf("DEBUG: cmd: %v\n", cmd)

	ctx := context.Background()

	switch cmd.Type {
	case "create_order":
		// Convert from OrderItemCommand to SimpleItem directly
		items := make([]SimpleItem, len(cmd.OrderItems))
		for i, item := range cmd.OrderItems {
			items[i] = SimpleItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			}
		}
		// // Convert order items from interface to SimpleItem
		// var items []SimpleItem
		// itemsData, err := json.Marshal(cmd.OrderItems)
		// fmt.Println("itemsData", string(itemsData))
		// if err != nil {
		// 	return fmt.Errorf("failed to marshal order items: %w", err)
		// }

		// if err := json.Unmarshal(itemsData, &items); err != nil {
		// 	return fmt.Errorf("failed to unmarshal order items: %w", err)
		// }

		notes := ""
		if notesVal, ok := cmd.AdditionalData["notes"]; ok {
			if notesStr, ok := notesVal.(string); ok {
				notes = notesStr
			}
		}

		// Call the underlying service to create the order
		fmt.Printf("DEBUG: cmd.CustomerID: %v\n", cmd.CustomerID)
		fmt.Printf("DEBUG: cmd.MerchantID: %v\n", cmd.MerchantID)
		fmt.Printf("DEBUG: items: %v\n", items)
		fmt.Printf("DEBUG: notes: %v\n", notes)
		order, err := s.orderService.CreateOrder(ctx, cmd.CustomerID, cmd.MerchantID, items, notes)
		fmt.Printf("DEBUG: Order created: %v\n", order)
		if err != nil {
			return nil, fmt.Errorf("failed to create order: %w", err)
		}

		// Store the order ID in the command for reference
		cmd.OrderID = order.ID
		createdOrder = order
		fmt.Printf("DEBUG: cmd.OrderID: %v\n", cmd.OrderID)

	case "update_order_status":
		// Get the status from additional data
		statusStr, ok := cmd.AdditionalData["status"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid status in update_order_status command")
		}

		// Convert string to OrderStatus
		status := domain.OrderStatus(statusStr)

		// Call the underlying service to update the order status
		if err := s.orderService.UpdateStatus(ctx, cmd.OrderID, status); err != nil {
			return nil, fmt.Errorf("failed to update order status: %w", err)
		}

		// For status updates, we don't need to return the order
		return nil, nil

	case "update_order":
		// Handle the update_order command too
		statusStr, _ := cmd.AdditionalData["status"].(string)
		notesStr, _ := cmd.AdditionalData["notes"].(string)

		if err := s.orderService.UpdateOrder(ctx, cmd.OrderID, statusStr, notesStr); err != nil {
			return nil, fmt.Errorf("failed to update order: %w", err)
		}

		return nil, nil

	case "update_inventory":
		// Process inventory updates without orders
		// This would be implemented similarly to create_order

	default:
		return nil, fmt.Errorf("unknown command type: %s", cmd.Type)
	}

	return createdOrder, nil
}

// processAppliedCommands listens for applied log entries and processes them
func (s *RaftOrderService) processAppliedCommands() {
	for entry := range s.applyCh {
		// Log that we received a command for auditing
		s.logger.Printf("Applied command at index %d, term %d", entry.Index, entry.Term)

		// Apply the command directly and store the result
		order, err := s.applyCommand(entry.Command)
		if err != nil {
			s.logger.Printf("Error applying command: %v", err)
			continue
		}

		// If it's an order creation command and the order was created successfully, store the result
		if order != nil {
			s.resultMapLock.Lock()
			s.resultMap[entry.Index] = order
			s.resultMapLock.Unlock()
		}
	}
}

// Other methods that the OrderService has, such as GetByID, UpdateStatus, etc.
// These methods can go directly to the underlying OrderService since they don't
// affect distributed state

func (s *RaftOrderService) GetByID(ctx context.Context, id uint) (*domain.Order, []domain.OrderItem, error) {
	return s.orderService.GetByID(ctx, id)
}

func (s *RaftOrderService) ListByCustomer(ctx context.Context, cid uint) ([]*domain.Order, error) {
	return s.orderService.ListByCustomer(ctx, cid)
}

func (s *RaftOrderService) ListByMerchant(ctx context.Context, mid uint) ([]*domain.Order, error) {
	return s.orderService.ListByMerchant(ctx, mid)
}

func (s *RaftOrderService) UpdateStatus(ctx context.Context, id uint, st domain.OrderStatus) error {
	// For status changes that affect inventory (like cancellations), use Raft
	// Otherwise, go directly to the underlying service
	if st == domain.OrderStatusCancelled || st == domain.OrderStatusCompleted {
		cmd := raft.OrderCommand{
			Type:    "update_order_status",
			OrderID: id,
			AdditionalData: map[string]interface{}{
				"status": string(st),
			},
		}

		_, err := s.raftNode.Submit(cmd)
		return err
	}

	return s.orderService.UpdateStatus(ctx, id, st)
}

func (s *RaftOrderService) UpdateOrder(ctx context.Context, id uint, status string, notes string) error {
	// For status changes that affect inventory, use Raft
	if status == string(domain.OrderStatusCancelled) {
		cmd := raft.OrderCommand{
			Type:    "update_order",
			OrderID: id,
			AdditionalData: map[string]interface{}{
				"status": status,
				"notes":  notes,
			},
		}

		_, err := s.raftNode.Submit(cmd)
		return err
	}

	return s.orderService.UpdateOrder(ctx, id, status, notes)
}

func (s *RaftOrderService) CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error) {
	return s.orderService.CheckProductsAvailability(ctx, productIDs)
}

// GetRaftNode returns the underlying Raft node
func (s *RaftOrderService) GetRaftNode() *raft.RaftNode {
	return s.raftNode
}

func (s *RaftOrderService) updateLastApplied(index uint64) {
	s.raftNode.UpdateLastApplied(index)
}

// CleanupResults cleans up the result map
func (s *RaftOrderService) cleanupResults() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.resultMapLock.Lock()
		// Clean up based on time or maximum number
		// Here we simplify the cleanup, in actual use, it should be more refined
		if len(s.resultMap) > 1000 {
			s.resultMap = make(map[uint64]*domain.Order)
		}
		s.resultMapLock.Unlock()
	}
}

// DeleteOrder deletes an order with Raft consensus
func (s *RaftOrderService) DeleteOrder(ctx context.Context, id uint) error {
	// Get the order first to check its status
	order, _, err := s.orderService.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// If order is pending, we need to cancel it first (which will use Raft)
	if order.Status == domain.OrderStatusPending {
		fmt.Printf("DEBUG: Order is pending, need to cancel\n")

		if err := s.UpdateStatus(ctx, id, domain.OrderStatusCancelled); err != nil {
			return fmt.Errorf("failed to cancel order before deletion: %w", err)
		}

	}

	// No need to delete order with Raft, just use the underlying service
	return s.orderService.DeleteOrder(ctx, id)

}
