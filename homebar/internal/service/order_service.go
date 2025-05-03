package service

import (
	"context"
	"time"

	"errors"
	"fmt"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository"
	"github.com/kexincchen/homebar/internal/repository/postgres"
)

type OrderService struct {
	orderRepo         repository.OrderRepository
	productRepo       repository.ProductRepository
	ingredientService *IngredientService
	inventoryRepo     *postgres.InventoryRepository
}

func NewOrderService(or repository.OrderRepository, pr repository.ProductRepository, ingredientService *IngredientService, inventoryRepo *postgres.InventoryRepository) *OrderService {
	return &OrderService{or, pr, ingredientService, inventoryRepo}
}

type SimpleItem struct {
	ProductID uint
	Quantity  int
	Price     float64
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	customerID, merchantID uint,
	items []SimpleItem,
	notes string,
) (*domain.Order, error) {
	fmt.Printf("DEBUG: Creating order with customerID=%d, merchantID=%d\n", customerID, merchantID)
	fmt.Printf("DEBUG: Items: %v\n", items)

	var (
		total  float64
		models []domain.OrderItem
	)
	for _, it := range items {
		price := it.Price
		if price == 0 {
			p, err := s.productRepo.GetByID(ctx, it.ProductID)
			fmt.Printf("DEBUG: Product: %v\n", p)
			if err != nil {
				return nil, err
			}
			price = p.Price
		}
		total += price * float64(it.Quantity)
		models = append(models, domain.OrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     price,
		})
	}

	now := time.Now()
	order := &domain.Order{
		CustomerID:  customerID,
		MerchantID:  merchantID,
		TotalAmount: total,
		Status:      domain.OrderStatusPending,
		Notes:       notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Convert models to slice of pointers for HasSufficientInventoryForOrder
	modelPtrs := make([]*domain.OrderItem, len(models))
	for i := range models {
		modelPtrs[i] = &models[i]
	}

	// Check if we have enough inventory for this order
	fmt.Printf("DEBUG: Checking inventory for %d items\n", len(modelPtrs))
	hasInventory, err := s.ingredientService.HasSufficientInventoryForOrder(ctx, modelPtrs)
	fmt.Printf("DEBUG: Has inventory: %v\n", hasInventory)
	if err != nil {
		return nil, err
	}

	if !hasInventory {
		return nil, errors.New("insufficient ingredients inventory for this order")
	}

	fmt.Printf("DEBUG: Creating order in database\n")
	if err := s.orderRepo.Create(ctx, order, models); err != nil {
		fmt.Printf("DEBUG: Error creating order in database: %v\n", err)
		return nil, err
	}

	// The inventory is already updated by the HasSufficientInventoryForOrder method
	// which locks and reduces the inventory when successful
	fmt.Printf("DEBUG: Order created in database\n")
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uint) (*domain.Order, []domain.OrderItem, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *OrderService) ListByCustomer(ctx context.Context, cid uint) ([]*domain.Order, error) {
	return s.orderRepo.GetByCustomer(ctx, cid)
}

func (s *OrderService) ListByMerchant(ctx context.Context, mid uint) ([]*domain.Order, error) {
	return s.orderRepo.GetByMerchant(ctx, mid)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id uint, status domain.OrderStatus) error {
	// First get the current order status
	order, _, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check for valid status transitions
	if !isValidStatusTransition(order.Status, status) {
		return errors.New("invalid status transition")
	}

	// Start transaction
	tx, err := s.orderRepo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update the order status
	err = s.orderRepo.UpdateStatus(ctx, tx, id, status)
	if err != nil {
		return err
	}

	// Handle inventory based on status change - ONLY for status changes to cancelled
	if status == domain.OrderStatusCancelled && order.Status == domain.OrderStatusPending {
		// For cancelled orders, restore the inventory
		err = s.ingredientService.CancelOrderInventory(ctx, id)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(current, new domain.OrderStatus) bool {
	switch current {
	case domain.OrderStatusPending:
		// Pending orders can be completed or cancelled
		return new == domain.OrderStatusCompleted || new == domain.OrderStatusCancelled
	case domain.OrderStatusCompleted, domain.OrderStatusCancelled:
		// Completed or cancelled orders cannot change status
		return false
	default:
		return false
	}
}

// UpdateOrder updates an order's details
func (s *OrderService) UpdateOrder(ctx context.Context, id uint, status string, notes string) error {
	// Get the existing order
	order, _, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := s.orderRepo.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Track if we need to update the order
	needsUpdate := false

	// Handle status change if provided
	statusChanged := false
	newStatus := order.Status // Default to current status

	if status != "" && string(order.Status) != status {
		newStatus = domain.OrderStatus(status)
		if !isValidStatusTransition(order.Status, newStatus) {
			return errors.New("invalid status transition")
		}

		statusChanged = true
		needsUpdate = true
	}

	// Update notes if provided
	notesChanged := false
	if notes != "" && notes != order.Notes {
		order.Notes = notes
		notesChanged = true
		needsUpdate = true
	}

	// If nothing changed, return early
	if !needsUpdate {
		return nil
	}

	// Update the order in the database
	if statusChanged {
		err = s.orderRepo.UpdateStatus(ctx, tx, id, newStatus)
	} else if notesChanged {
		err = s.orderRepo.Update(ctx, tx, order)
	}

	if err != nil {
		return err
	}

	// Handle inventory ONLY in UpdateStatus method
	// Remove inventory handling from here to avoid duplication

	// Commit the transaction
	return tx.Commit()
}

// func (s *OrderService) verifyInventory(
// 	ctx context.Context,
// 	merchantID uint,
// 	items []SimpleItem,
// ) error {
// 	return nil
// }

// func (s *OrderService) updateInventory(
// 	ctx context.Context,
// 	merchantID uint,
// 	items []SimpleItem,
// 	orderID uint,
// ) error {
// 	return nil
// }

func (s *OrderService) CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error) {
	return s.ingredientService.CheckProductsAvailability(ctx, productIDs)
}
