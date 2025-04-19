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

func NewOrderService(or repository.OrderRepository, pr repository.ProductRepository, ir repository.InventoryRepository) *OrderService {
	return &OrderService{or, pr, ir}
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

	if err := s.verifyInventory(ctx, merchantID, items); err != nil {
		return nil, err
	}

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

	if err := s.orderRepo.Create(ctx, order, models); err != nil {
		return nil, err
	}
	if err := s.updateInventory(ctx, merchantID, items, order.ID); err != nil {
		return nil, err
	}
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
