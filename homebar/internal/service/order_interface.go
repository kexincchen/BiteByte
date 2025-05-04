package service

import (
	"context"

	"github.com/kexincchen/homebar/internal/domain"
)

// OrderServiceInterface defines methods that both OrderService and RaftOrderService implement
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, customerID, merchantID uint, items []SimpleItem, notes string) (*domain.Order, error)
	GetByID(ctx context.Context, id uint) (*domain.Order, []domain.OrderItem, error)
	ListByCustomer(ctx context.Context, cid uint) ([]*domain.Order, error)
	ListByMerchant(ctx context.Context, mid uint) ([]*domain.Order, error)
	UpdateStatus(ctx context.Context, id uint, st domain.OrderStatus) error
	UpdateOrder(ctx context.Context, id uint, status string, notes string) error
	CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error)
	DeleteOrder(ctx context.Context, id uint) error
}
