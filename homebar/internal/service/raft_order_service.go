package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/raft"
)

// RaftOrderService wraps OrderService to provide distributed consensus
type RaftOrderService struct {
	orderService *OrderService
	raftNode     *raft.RaftNode
	applyCh      chan raft.LogEntry
	nodeID       string
	isLeader     bool
	logger       *log.Logger
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
	}

	// Create the Raft node
	raftNode := raft.NewRaftNode(
		nodeID,
		peerIDs,
		peerAddrs,
		applyCh,
		service.applyCommand,
		logger,
	)

	service.raftNode = raftNode

	// Start processing applied commands
	go service.processAppliedCommands()

	return service, nil
}

// Start initializes and starts the Raft node
func (s *RaftOrderService) Start(ctx context.Context) error {
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
	orderItems := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		orderItems = append(orderItems, map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
			"price":      item.Price,
		})
	}

	cmd := raft.OrderCommand{
		Type:       "create_order",
		CustomerID: customerID,
		MerchantID: merchantID,
		OrderItems: orderItems,
		AdditionalData: map[string]interface{}{
			"notes": notes,
		},
	}

	// Submit the command to Raft
	index, err := s.raftNode.Submit(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to submit order to Raft: %w", err)
	}

	// Wait for the command to be applied
	// In a real system, we would have a more sophisticated waiting mechanism
	// This is a simplified approach
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if the command has been applied
			s.updateLastApplied(index)

			// Retrieve the order that was created
			// This assumes the command was successfully applied
			// In a real system, we would track command results more carefully
			order, _, err := s.orderService.GetByID(ctx, cmd.OrderID)
			return order, err

		case <-timeout:
			return nil, errors.New("timeout waiting for order creation")

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// applyCommand applies a Raft command to the state machine
func (s *RaftOrderService) applyCommand(cmdInterface interface{}) error {
	// Convert the interface to an OrderCommand
	cmdBytes, err := json.Marshal(cmdInterface)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	var cmd raft.OrderCommand
	if err := json.Unmarshal(cmdBytes, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal command: %w", err)
	}

	ctx := context.Background()

	switch cmd.Type {
	case "create_order":
		// Convert order items from interface to SimpleItem
		var items []SimpleItem
		itemsData, err := json.Marshal(cmd.OrderItems)
		if err != nil {
			return fmt.Errorf("failed to marshal order items: %w", err)
		}

		if err := json.Unmarshal(itemsData, &items); err != nil {
			return fmt.Errorf("failed to unmarshal order items: %w", err)
		}

		notes := ""
		if notesVal, ok := cmd.AdditionalData["notes"]; ok {
			if notesStr, ok := notesVal.(string); ok {
				notes = notesStr
			}
		}

		// Call the underlying service to create the order
		order, err := s.orderService.CreateOrder(ctx, cmd.CustomerID, cmd.MerchantID, items, notes)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Store the order ID in the command for reference
		cmd.OrderID = order.ID

	case "update_inventory":
		// Process inventory updates without orders
		// This would be implemented similarly to create_order

	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}

	return nil
}

// processAppliedCommands listens for applied log entries and processes them
func (s *RaftOrderService) processAppliedCommands() {
	for entry := range s.applyCh {
		// Log that we received a command for auditing
		s.logger.Printf("Applied command at index %d, term %d", entry.Index, entry.Term)

		// Commands are applied directly by the applyCommand function
		// This loop exists mainly for logging and monitoring
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
	if st == domain.OrderStatusCancelled {
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
