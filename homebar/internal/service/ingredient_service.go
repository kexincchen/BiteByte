package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/raft"
	"github.com/kexincchen/homebar/internal/repository/postgres"
)

type IngredientService struct {
	ingredientRepo        *postgres.IngredientRepository
	productIngredientRepo *postgres.ProductIngredientRepository
	inventoryRepo         *postgres.InventoryRepository
	db                    *sql.DB
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
		db:                    ingredientRepo.GetDB(),
	}
}

func (s *IngredientService) CreateIngredient(ctx context.Context, ingredient *domain.Ingredient) (*domain.Ingredient, error) {
	return s.ingredientRepo.Create(ctx, ingredient)
}

func (s *IngredientService) GetIngredientByID(ctx context.Context, id int64) (*domain.Ingredient, error) {
	fmt.Println("Getting ingredient by ID: ", id)
	return s.ingredientRepo.GetByID(ctx, id)
}

func (s *IngredientService) GetIngredientsByMerchant(ctx context.Context, merchantID int64) ([]*domain.Ingredient, error) {
	return s.ingredientRepo.GetByMerchant(ctx, merchantID)
}

func (s *IngredientService) UpdateIngredient(ctx context.Context, ingredient *domain.Ingredient) error {
	return s.ingredientRepo.Update(ctx, ingredient)
}

func (s *IngredientService) DeleteIngredient(ctx context.Context, id int64) error {
	fmt.Println("Deleting ingredient: ", id)
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

// HasSufficientInventoryForOrderWithRaft checks if there is sufficient inventory for an order using Raft
func (s *IngredientService) HasSufficientInventoryForOrderWithRaft(
	ctx context.Context,
	raftNode *raft.RaftNode,
	orderItems []*domain.OrderItem,
) (bool, error) {
	// First check locally if we have enough inventory
	hasInventory, err := s.HasSufficientInventoryForOrder(ctx, orderItems)
	if err != nil || !hasInventory {
		return hasInventory, err
	}

	// If we do, create a Raft command to reserve the inventory
	var orderItemsData []map[string]interface{}
	for _, item := range orderItems {
		orderItemsData = append(orderItemsData, map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
		})
	}

	cmd := raft.OrderCommand{
		Type:       "reserve_inventory",
		OrderItems: orderItemsData,
	}

	// Submit to Raft to achieve consensus
	_, err = raftNode.Submit(cmd)
	if err != nil {
		return false, fmt.Errorf("failed to achieve consensus on inventory reservation: %w", err)
	}

	// If the command was accepted by Raft, the inventory is reserved
	return true, nil
}

// ReserveInventoryCommand processes a command to reserve inventory
func (s *IngredientService) ReserveInventoryCommand(
	ctx context.Context,
	cmd raft.OrderCommand,
) error {
	// Convert order items from the command
	var orderItems []*domain.OrderItem
	itemsData, err := json.Marshal(cmd.OrderItems)
	if err != nil {
		return fmt.Errorf("failed to marshal order items: %w", err)
	}

	if err := json.Unmarshal(itemsData, &orderItems); err != nil {
		return fmt.Errorf("failed to unmarshal order items: %w", err)
	}

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reserve inventory for each order item
	for _, item := range orderItems {
		// Get the product ingredients
		ingredients, err := s.productIngredientRepo.GetProductIngredients(ctx, int64(item.ProductID))
		if err != nil {
			return err
		}

		// Lock and update each ingredient
		for _, prodIngredient := range ingredients {
			// Lock the ingredient row - use GetByID with a transaction lock instead
			var ingredient *domain.Ingredient
			err := tx.QueryRowContext(
				ctx,
				"SELECT * FROM ingredients WHERE id = $1 FOR UPDATE",
				prodIngredient.IngredientID,
			).Scan(&ingredient.ID, &ingredient.MerchantID, &ingredient.Name, &ingredient.Description,
				&ingredient.Unit, &ingredient.Quantity, &ingredient.CreatedAt, &ingredient.UpdatedAt)
			if err != nil {
				return err
			}

			// Calculate required quantity
			requiredQty := prodIngredient.Quantity * float64(item.Quantity)

			// Check if we have enough
			if ingredient.Quantity < requiredQty {
				return fmt.Errorf("insufficient quantity of ingredient %d", prodIngredient.IngredientID)
			}

			// Update the ingredient quantity
			_, err = tx.ExecContext(
				ctx,
				"UPDATE ingredients SET quantity = quantity - $1 WHERE id = $2",
				requiredQty,
				ingredient.ID,
			)
			if err != nil {
				return err
			}

			// Record the reservation
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO inventory_reservations 
				(order_id, ingredient_id, quantity, status, created_at, updated_at)
				VALUES ($1, $2, $3, 'reserved', NOW(), NOW())`,
				cmd.OrderID, ingredient.ID, requiredQty,
			)
			if err != nil {
				return err
			}
		}
	}

	// Commit the transaction
	return tx.Commit()
}
