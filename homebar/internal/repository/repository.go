package repository

import (
	"context"
	"database/sql"

	"github.com/kexincchen/homebar/internal/repository/postgres"

	"github.com/kexincchen/homebar/internal/domain"
)

func NewProductRepository(db *sql.DB) ProductRepository {
	return postgres.NewProductRepository(db)
}

func NewUserRepository(db *sql.DB) UserRepository {
	return postgres.NewUserRepository(db)
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return postgres.NewOrderRepository(db)
}

func NewMerchantRepository(db *sql.DB) MerchantRepository {
	return postgres.NewMerchantRepository(db)
}

func NewCustomerRepository(db *sql.DB) CustomerRepository {
	return postgres.NewCustomerRepository(db)
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error
}

type ProductRepository interface {
	Create(ctx context.Context, p *domain.Product) error
	GetByID(ctx context.Context, id uint) (*domain.Product, error)
	Update(ctx context.Context, p *domain.Product) error
	Delete(ctx context.Context, id uint) error
	GetAll(ctx context.Context) ([]*domain.Product, error)
	GetByMerchant(ctx context.Context, merchantID uint) ([]*domain.Product, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order, items []domain.OrderItem) error
	GetByID(ctx context.Context, id uint) (*domain.Order, []domain.OrderItem, error)
	GetByCustomer(ctx context.Context, customerID uint) ([]*domain.Order, error)
	GetByMerchant(ctx context.Context, merchantID uint) ([]*domain.Order, error)
	UpdateStatus(ctx context.Context, orderID uint, status domain.OrderStatus) error
	UpdateOrder(ctx context.Context, o *domain.Order) error
}

type InventoryRepository interface {
	GetByMerchantAndIngredient(ctx context.Context, merchantID, ingredientID uint) (*domain.Inventory, error)
	UpdateQuantity(ctx context.Context, inventoryID uint, quantityChange float64) error
	LogTransaction(ctx context.Context, transaction *domain.InventoryTransaction) error
}

type MerchantRepository interface {
	Create(ctx context.Context, m *domain.Merchant) error
	GetByID(ctx context.Context, id uint) (*domain.Merchant, error)
	GetByUserID(ctx context.Context, userID uint) (*domain.Merchant, error)
	GetByUsername(ctx context.Context, username string) (*domain.Merchant, error)
	List(ctx context.Context) ([]*domain.Merchant, error)
	Update(ctx context.Context, m *domain.Merchant) error
	Delete(ctx context.Context, id uint) error
}

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByUserID(ctx context.Context, userID uint) (*domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, userID uint) error
}
