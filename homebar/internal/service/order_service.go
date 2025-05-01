package service

import (
	"context"
	"time"

	"errors"

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
	var (
		total  float64
		models []domain.OrderItem
	)
	for _, it := range items {
		price := it.Price
		if price == 0 {
			p, err := s.productRepo.GetByID(ctx, it.ProductID)
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
	hasInventory, err := s.ingredientService.HasSufficientInventoryForOrder(ctx, modelPtrs)
	if err != nil {
		return nil, err
	}

	if !hasInventory {
		return nil, errors.New("insufficient ingredients inventory for this order")
	}

	if err := s.orderRepo.Create(ctx, order, models); err != nil {
		return nil, err
	}

	// The inventory is already updated by the HasSufficientInventoryForOrder method
	// which locks and reduces the inventory when successful

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

func (s *OrderService) UpdateStatus(ctx context.Context, id uint, st domain.OrderStatus) error {
	return s.orderRepo.UpdateStatus(ctx, id, st)
}

func (s *OrderService) UpdateOrder(ctx context.Context, id uint, status string, notes string) error {
	// Get the existing order
	order, _, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update fields that are provided
	if status != "" {
		order.Status = domain.OrderStatus(status)
	}

	if notes != "" {
		order.Notes = notes
	}

	// Update the order in the database
	return s.orderRepo.UpdateOrder(ctx, order)
}

func (s *OrderService) verifyInventory(
	ctx context.Context,
	merchantID uint,
	items []SimpleItem,
) error {
	return nil
}

func (s *OrderService) updateInventory(
	ctx context.Context,
	merchantID uint,
	items []SimpleItem,
	orderID uint,
) error {
	return nil
}

func (s *OrderService) CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error) {
	return s.ingredientService.CheckProductsAvailability(ctx, productIDs)
}
