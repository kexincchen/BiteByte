package service

import (
	"context"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository"
)

type OrderService struct {
	orderRepo     repository.OrderRepository
	productRepo   repository.ProductRepository
	inventoryRepo repository.InventoryRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
	inventoryRepo repository.InventoryRepository,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, customerID uint, merchantID uint,
	items []struct {
		ProductID uint
		Quantity  int
	}, notes string) (*domain.Order, error) {

	// Verify inventory before creating order
	if err := s.verifyInventory(ctx, merchantID, items); err != nil {
		return nil, err
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		totalAmount += product.Price * float64(item.Quantity)
	}

	// Create order
	now := time.Now()
	order := &domain.Order{
		CustomerID:  customerID,
		MerchantID:  merchantID,
		TotalAmount: totalAmount,
		Status:      domain.OrderStatusPending,
		Notes:       notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Update inventory
	// This would be better in a transaction
	if err := s.updateInventory(ctx, merchantID, items, order.ID); err != nil {
		// Should rollback the order creation here in a real implementation
		return nil, err
	}

	return order, nil
}

func (s *OrderService) verifyInventory(ctx context.Context, merchantID uint,
	items []struct {
		ProductID uint
		Quantity  int
	}) error {
	// Implementation would check if there are enough ingredients for all products
	return nil
}

func (s *OrderService) updateInventory(ctx context.Context, merchantID uint,
	items []struct {
		ProductID uint
		Quantity  int
	}, orderID uint) error {
	// Implementation would reduce inventory quantities based on order items
	return nil
}
