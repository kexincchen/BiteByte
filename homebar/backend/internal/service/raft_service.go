package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/raft"
)

// RaftService wraps OrderService to provide distributed consensus
type RaftService struct {
	orderService        *OrderService
	ingredientService   *IngredientService
	raftNode            *raft.RaftNode
	applyCh             chan raft.LogEntry
	nodeID              string
	isLeader            bool
	orderResultMap      map[uint64]*domain.Order
	ingredientResultMap map[uint64]*domain.Ingredient
	resultMapLock       sync.Mutex
}

// NewRaftService creates a new Raft-enabled order service
func NewRaftService(
	orderService *OrderService,
	ingredientService *IngredientService,
	nodeID string,
	peerIDs []string,
	peerAddrs map[string]string,
) (*RaftService, error) {
	raftLogger := log.With().
        Str("component", "raft").
        Str("node_id", nodeID).
        Logger()

	// Initialize the Raft node
    raftLogger.Info().Msg("Initializing Raft service")

	// Create the apply channel
	applyCh := make(chan raft.LogEntry, raft.MaxLogEntriesBuffer)

	service := &RaftService{
		orderService:        orderService,
		ingredientService:   ingredientService,
		applyCh:             applyCh,
		nodeID:              nodeID,
		isLeader:            false,
		orderResultMap:      make(map[uint64]*domain.Order),
		ingredientResultMap: make(map[uint64]*domain.Ingredient),
	}

	// Create the Raft node
	raftNode := raft.NewRaftNode(
		nodeID,
		peerIDs,
		peerAddrs,
		applyCh,
		func(cmd interface{}) error {
			_, _, err := service.applyCommand(cmd)
			return err
		},
	)

	service.raftNode = raftNode

	// Start processing applied commands
	go service.processAppliedCommands()

	return service, nil
}

// Start initializes and starts the Raft node
func (s *RaftService) Start(ctx context.Context) error {
	// Start the cleanup goroutine
	go s.cleanupResults()

	// Start the Raft node
	return s.raftNode.Start(ctx)
}

// CreateOrder creates a new order with Raft consensus
func (s *RaftService) CreateOrder(
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
	// Submit the command to Raft
	index, err := s.raftNode.Submit(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to submit order to Raft: %w", err)
	}

	// Wait for the command to be applied
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Log performance metrics
			log.Info().
			Uint("customer_id", customerID).
			Uint("merchant_id", merchantID).
			Int("item_count", len(items)).
			Msg("Order created successfully")

			// Check if the command has been applied
			s.updateLastApplied(index)
			s.resultMapLock.Lock()
			order, exists := s.orderResultMap[index]
			if exists {
				// Delete the result from the cache to avoid memory leaks
				delete(s.orderResultMap, index)
				s.resultMapLock.Unlock()
				return order, nil
			}
			s.resultMapLock.Unlock()

		case <-timeout:
			return nil, errors.New("timeout waiting for order creation")

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// applyCommand applies a Raft command to the state machine
func (s *RaftService) applyCommand(cmdInterface interface{}) (*domain.Order, *domain.Ingredient, error) {
	var createdOrder *domain.Order = nil
	var createdIngredient *domain.Ingredient = nil

	// Convert the interface to an OrderCommand
	cmdBytes, err := json.Marshal(cmdInterface)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	var cmd raft.OrderCommand
	if err := json.Unmarshal(cmdBytes, &cmd); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}

	if s.nodeID != s.raftNode.LeaderID() {
		log.Printf("[Follower-%s] skip %s (already done by leader %s)",
			s.nodeID, cmd.Type, s.raftNode.LeaderID())
		return nil, nil, nil
	}

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

		notes := ""
		if notesVal, ok := cmd.AdditionalData["notes"]; ok {
			if notesStr, ok := notesVal.(string); ok {
				notes = notesStr
			}
		}

		// Call the underlying service to create the order
		order, err := s.orderService.CreateOrder(ctx, cmd.CustomerID, cmd.MerchantID, items, notes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create order: %w", err)
		}

		// Store the order ID in the command for reference
		cmd.OrderID = order.ID
		createdOrder = order

	case "update_order_status":
		// Get the status from additional data
		statusStr, ok := cmd.AdditionalData["status"].(string)
		if !ok {
			return nil, nil, fmt.Errorf("invalid status in update_order_status command")
		}

		// Convert string to OrderStatus
		status := domain.OrderStatus(statusStr)

		// Call the underlying service to update the order status
		if err := s.orderService.UpdateStatus(ctx, cmd.OrderID, status); err != nil {
			return nil, nil, fmt.Errorf("failed to update order status: %w", err)
		}

		// For status updates, we don't need to return the order
		return nil, nil, nil

	case "update_order":
		// Handle the update_order command too
		statusStr, _ := cmd.AdditionalData["status"].(string)
		notesStr, _ := cmd.AdditionalData["notes"].(string)

		if err := s.orderService.UpdateOrder(ctx, cmd.OrderID, statusStr, notesStr); err != nil {
			return nil, nil, fmt.Errorf("failed to update order: %w", err)
		}

		return nil, nil, nil

	case "create_ingredient":
		// Extract ingredient data from command
		var ingredient domain.Ingredient
		ingredientData, err := json.Marshal(cmd.AdditionalData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal ingredient data: %w", err)
		}

		if err := json.Unmarshal(ingredientData, &ingredient); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal ingredient data: %w", err)
		}

		// Call the underlying service to create the ingredient
		result, err := s.ingredientService.CreateIngredient(ctx, &ingredient)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create ingredient: %w", err)
		}

		createdIngredient = result

	case "update_ingredient":
		// Extract ingredient data from command
		var ingredient domain.Ingredient
		ingredientData, err := json.Marshal(cmd.AdditionalData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal ingredient data: %w", err)
		}

		if err := json.Unmarshal(ingredientData, &ingredient); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal ingredient data: %w", err)
		}

		// Call the underlying service to update the ingredient
		if err := s.ingredientService.UpdateIngredient(ctx, &ingredient); err != nil {
			return nil, nil, fmt.Errorf("failed to update ingredient: %w", err)
		}

	case "delete_ingredient":
		// Get the ingredient ID
		idFloat, ok := cmd.AdditionalData["id"].(float64)
		if !ok {
			return nil, nil, fmt.Errorf("invalid ingredient ID in delete_ingredient command")
		}

		id := int64(idFloat)

		// Call the underlying service to delete the ingredient
		if err := s.ingredientService.DeleteIngredient(ctx, id); err != nil {
			return nil, nil, fmt.Errorf("failed to delete ingredient: %w", err)
		}

	default:
		return nil, nil, fmt.Errorf("unknown command type: %s", cmd.Type)
	}

	return createdOrder, createdIngredient, nil
}

// processAppliedCommands listens for applied log entries and processes them
func (s *RaftService) processAppliedCommands() {
	for entry := range s.applyCh {
		// Log that we received a command for auditing
		log.Printf("Applied command at index %d, term %d", entry.Index, entry.Term)

		// Apply the command directly and store the result
		order, ingredient, err := s.applyCommand(entry.Command)
		if err != nil {
			log.Printf("Error applying command: %v", err)
			continue
		}

		s.resultMapLock.Lock()

		// If it's an order creation command and the order was created successfully
		if order != nil {
			s.orderResultMap[entry.Index] = order
		}

		// If it's an ingredient creation command and was successful
		if ingredient != nil {
			s.ingredientResultMap[entry.Index] = ingredient
		}

		s.resultMapLock.Unlock()
	}
}

func (s *RaftService) UpdateOrder(ctx context.Context, id uint, status string, notes string) error {
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

func (s *RaftService) UpdateStatus(ctx context.Context, id uint, st domain.OrderStatus) error {
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

// DeleteOrder deletes an order with Raft consensus
func (s *RaftService) DeleteOrder(ctx context.Context, id uint) error {
	// Get the order first to check its status
	order, _, err := s.orderService.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// If order is pending, we need to cancel it first (which will use Raft)
	if order.Status == domain.OrderStatusPending {

		if err := s.UpdateStatus(ctx, id, domain.OrderStatusCancelled); err != nil {
			return fmt.Errorf("failed to cancel order before deletion: %w", err)
		}

	}

	// No need to delete order with Raft, just use the underlying service
	return s.orderService.DeleteOrder(ctx, id)

}

// CreateIngredient creates a new ingredient with Raft consensus
func (s *RaftService) CreateIngredient(ctx context.Context, ingredient *domain.Ingredient) (*domain.Ingredient, error) {
	// Convert ingredient to a map for the command
	ingredientData, err := json.Marshal(ingredient)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ingredient: %w", err)
	}

	var ingredientMap map[string]interface{}
	if err := json.Unmarshal(ingredientData, &ingredientMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ingredient: %w", err)
	}

	// Prepare the ingredient command
	cmd := raft.OrderCommand{
		Type:           "create_ingredient",
		AdditionalData: ingredientMap,
	}

	// Submit the command to Raft
	index, err := s.raftNode.Submit(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to submit ingredient creation to Raft: %w", err)
	}

	// Wait for the command to be applied
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateLastApplied(index)
			s.resultMapLock.Lock()
			result, exists := s.ingredientResultMap[index]
			if exists {
				delete(s.ingredientResultMap, index)
				s.resultMapLock.Unlock()
				return result, nil
			}
			s.resultMapLock.Unlock()
		case <-timeout:
			return nil, errors.New("timeout waiting for ingredient creation")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// DeleteIngredient deletes an ingredient with Raft consensus
func (s *RaftService) DeleteIngredient(ctx context.Context, id int64) error {
	// Prepare the ingredient command
	cmd := raft.OrderCommand{
		Type: "delete_ingredient",
		AdditionalData: map[string]interface{}{
			"id": id,
		},
	}

	// Submit the command to Raft
	_, err := s.raftNode.Submit(cmd)
	if err != nil {
		return fmt.Errorf("failed to submit ingredient deletion to Raft: %w", err)
	}

	// For deletions, we don't need to wait for a result
	return nil
}

// GetRaftNode returns the underlying Raft node
func (s *RaftService) GetRaftNode() *raft.RaftNode {
	return s.raftNode
}

func (s *RaftService) updateLastApplied(index uint64) {
	s.raftNode.UpdateLastApplied(index)
}

// CleanupResults cleans up the result map
func (s *RaftService) cleanupResults() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.resultMapLock.Lock()
		// Clean up based on time or maximum number
		// Here we simplify the cleanup, in actual use, it should be more refined
		if len(s.orderResultMap) > 1000 {
			s.orderResultMap = make(map[uint64]*domain.Order)
		}
		if len(s.ingredientResultMap) > 1000 {
			s.ingredientResultMap = make(map[uint64]*domain.Ingredient)
		}
		s.resultMapLock.Unlock()
	}
}

// UpdateIngredient updates an ingredient with Raft consensus
func (s *RaftService) UpdateIngredient(ctx context.Context, ingredient *domain.Ingredient) error {
	// Convert ingredient to a map for the command
	ingredientData, err := json.Marshal(ingredient)
	if err != nil {
		return fmt.Errorf("failed to marshal ingredient: %w", err)
	}

	var ingredientMap map[string]interface{}
	if err := json.Unmarshal(ingredientData, &ingredientMap); err != nil {
		return fmt.Errorf("failed to unmarshal ingredient: %w", err)
	}

	// Prepare the ingredient command
	cmd := raft.OrderCommand{
		Type:           "update_ingredient",
		AdditionalData: ingredientMap,
	}

	// Submit the command to Raft
	_, err = s.raftNode.Submit(cmd)
	if err != nil {
		return fmt.Errorf("failed to submit ingredient update to Raft: %w", err)
	}

	// For updates, we don't need to wait for a result
	return nil
}

// Other methods that the OrderService has, such as GetByID, UpdateStatus, etc.
// These methods can go directly to the underlying OrderService since they don't
// affect distributed state

func (s *RaftService) GetByID(ctx context.Context, id uint) (*domain.Order, []domain.OrderItem, error) {
	return s.orderService.GetByID(ctx, id)
}

func (s *RaftService) ListByCustomer(ctx context.Context, cid uint) ([]*domain.Order, error) {
	return s.orderService.ListByCustomer(ctx, cid)
}

func (s *RaftService) ListByMerchant(ctx context.Context, mid uint) ([]*domain.Order, error) {
	return s.orderService.ListByMerchant(ctx, mid)
}

func (s *RaftService) CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error) {
	return s.orderService.CheckProductsAvailability(ctx, productIDs)
}

// CheckProductAvailability checks if a product is available
func (s *RaftService) CheckProductAvailability(ctx context.Context, productID uint) (bool, error) {
	return s.ingredientService.CheckProductAvailability(ctx, productID)
}

// GetIngredientByID retrieves an ingredient by its ID
func (s *RaftService) GetIngredientByID(ctx context.Context, id int64) (*domain.Ingredient, error) {
	// This is a read operation, so we can delegate directly
	return s.ingredientService.GetIngredientByID(ctx, id)
}

// GetIngredientsByMerchant retrieves all ingredients for a merchant
func (s *RaftService) GetIngredientsByMerchant(ctx context.Context, merchantID int64) ([]*domain.Ingredient, error) {
	// This is a read operation, so we can delegate directly
	return s.ingredientService.GetIngredientsByMerchant(ctx, merchantID)
}

// GetInventorySummary retrieves inventory summary for a merchant
func (s *RaftService) GetInventorySummary(ctx context.Context, merchantID int64) (map[string]interface{}, error) {
	// This is a read operation, so we can delegate directly
	return s.ingredientService.GetInventorySummary(ctx, merchantID)
}

func init() {
    // Pretty console logging for development
    log.Logger = log.Output(zerolog.ConsoleWriter{
        Out:        os.Stdout,
        TimeFormat: time.RFC3339,
        NoColor:    false,
    })
    
    // Set global log level
    zerolog.SetGlobalLevel(zerolog.InfoLevel)
    
    // Enable caller information
    zerolog.CallerSkipFrameCount = 3
    log.Logger = log.With().Logger()
}