package service

import (
	"context"
	"database/sql"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository/postgres"
)

type IngredientService struct {
	ingredientRepo        *postgres.IngredientRepository
	productIngredientRepo *postgres.ProductIngredientRepository
	inventoryRepo         *postgres.InventoryRepository
}

func NewIngredientService(
	ingredientRepo *postgres.IngredientRepository,
	productIngredientRepo *postgres.ProductIngredientRepository,
	inventoryRepo *postgres.InventoryRepository,
) *IngredientService {
	return &IngredientService{
		ingredientRepo:        ingredientRepo,
		productIngredientRepo: productIngredientRepo,
		inventoryRepo:         inventoryRepo,
	}
}

func (s *IngredientService) CreateIngredient(ctx context.Context, ingredient *domain.Ingredient) (*domain.Ingredient, error) {
	return s.ingredientRepo.Create(ctx, ingredient)
}

func (s *IngredientService) GetIngredientByID(ctx context.Context, id int64) (*domain.Ingredient, error) {
	return s.ingredientRepo.GetByID(ctx, id)
}

func (s *IngredientService) GetIngredientsByMerchant(ctx context.Context, merchantID int64) ([]*domain.Ingredient, error) {
	return s.ingredientRepo.GetByMerchant(ctx, merchantID)
}

func (s *IngredientService) UpdateIngredient(ctx context.Context, ingredient *domain.Ingredient) error {
	return s.ingredientRepo.Update(ctx, ingredient)
}

func (s *IngredientService) DeleteIngredient(ctx context.Context, id int64) error {
	return s.ingredientRepo.Delete(ctx, id)
}

func (s *IngredientService) GetInventorySummary(ctx context.Context, merchantID int64) (map[string]interface{}, error) {
	return s.ingredientRepo.GetInventorySummary(ctx, merchantID)
}

func (s *IngredientService) HasSufficientInventoryForOrder(ctx context.Context, orderItems []*domain.OrderItem) (bool, error) {
	return s.ingredientRepo.LockInventoryForOrder(ctx, orderItems)
}

// CheckProductAvailability determines if a product is available based on its ingredients
func (s *IngredientService) CheckProductAvailability(ctx context.Context, productID uint) (bool, error) {
	// Get all ingredients required for this product
	ingredients, err := s.productIngredientRepo.GetProductIngredients(ctx, int64(productID))
	if err != nil {
		return false, err
	}

	// If product has no ingredients, consider it available
	if len(ingredients) == 0 {
		return true, nil
	}

	// Check if all ingredients are available in sufficient quantity
	for _, ingredient := range ingredients {
		// Get current inventory level for this ingredient
		inventory, err := s.inventoryRepo.GetByID(ctx, uint(ingredient.IngredientID))
		if err != nil {
			return false, err
		}

		// Check if there's enough inventory
		if inventory.Quantity < ingredient.Quantity {
			return false, nil
		}
	}

	return true, nil
}

// CheckProductsAvailability checks availability for multiple products at once
func (s *IngredientService) CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error) {
	return s.inventoryRepo.CheckProductsAvailability(ctx, productIDs)
}

// ReserveInventoryForOrder creates inventory reservations for an order
// This should be called within a transaction
func (s *IngredientService) ReserveInventoryForOrder(ctx context.Context, tx *sql.Tx, orderID uint, orderItems []*domain.OrderItem) error {
	// This function should:
	// 1. Calculate ingredients needed for all products in the order
	// 2. Lock the inventory rows (SELECT FOR UPDATE)
	// 3. Verify sufficient inventory
	// 4. Create reservation records
	// 5. Update inventory quantities

	// Get product IDs from order items
	var productIDs []uint
	for _, item := range orderItems {
		productIDs = append(productIDs, item.ProductID)
	}

	// Implementation will depend on your existing repository structure
	// This is intended to enhance your existing HasSufficientInventoryForOrder method

	return nil
}

// CompleteOrderInventory marks inventory reservations as completed
func (s *IngredientService) CompleteOrderInventory(ctx context.Context, tx *sql.Tx, orderID uint) error {
	// Mark all reservations for this order as 'completed'
	// No inventory quantity changes needed as they were already reduced during reservation
	return nil
}

// CancelOrderInventory returns reserved inventory
func (s *IngredientService) CancelOrderInventory(ctx context.Context, tx *sql.Tx, orderID uint) error {
	// 1. Find all 'reserved' reservations for this order
	// 2. Return quantities back to inventory
	// 3. Mark reservations as 'canceled'
	return nil
}
