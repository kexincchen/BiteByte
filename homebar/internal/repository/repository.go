package repository

import (
	"context"
	"database/sql"
	"github.com/kexincchen/homebar/internal/repository/postgres"

	"github.com/kexincchen/homebar/internal/domain"
)

// NewProductRepository returns a Postgres implementation that satisfies
// ProductRepository. main.go only needs to import the interface package.
func NewProductRepository(db *sql.DB) ProductRepository {
	return postgres.NewProductRepository(db)
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error
}

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id uint) (*domain.Product, error)
	GetByMerchant(ctx context.Context, merchantID uint) ([]*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id uint) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uint) (*domain.Order, error)
	GetByCustomer(ctx context.Context, customerID uint) ([]*domain.Order, error)
	GetByMerchant(ctx context.Context, merchantID uint) ([]*domain.Order, error)
	UpdateStatus(ctx context.Context, orderID uint, status domain.OrderStatus) error
}

type InventoryRepository interface {
	GetByMerchantAndIngredient(ctx context.Context, merchantID, ingredientID uint) (*domain.Inventory, error)
	UpdateQuantity(ctx context.Context, inventoryID uint, quantityChange float64) error
	LogTransaction(ctx context.Context, transaction *domain.InventoryTransaction) error
}
