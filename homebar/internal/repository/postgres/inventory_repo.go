package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
)

// InventoryRepository handles inventory operations
type InventoryRepository struct {
	db *sql.DB
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// GetByID retrieves inventory information for a specific ingredient
func (r *InventoryRepository) GetByID(ctx context.Context, id uint) (*domain.Ingredient, error) {
	query := `
		SELECT id, name, description, unit, current_quantity, created_at, updated_at 
		FROM inventory_items
		WHERE id = $1
	`
	
	var ingredient domain.Ingredient
	var createdAt, updatedAt time.Time
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ingredient.ID,
		&ingredient.Name,
		&ingredient.Description,
		&ingredient.Unit,
		&ingredient.Quantity,
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("inventory item with ID %d not found", id)
		}
		return nil, err
	}
	
	ingredient.CreatedAt = createdAt
	ingredient.UpdatedAt = updatedAt
	
	return &ingredient, nil
}

// GetByProductID retrieves all ingredients required for a product
func (r *InventoryRepository) GetByProductID(ctx context.Context, productID uint) ([]*domain.ProductIngredient, error) {
	query := `
		SELECT pi.id, pi.product_id, pi.ingredient_id, pi.quantity,
			i.name, i.unit
		FROM product_ingredients pi
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE pi.product_id = $1
	`
	
	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ingredients []*domain.ProductIngredient
	for rows.Next() {
		var pi domain.ProductIngredient
		var ingredientName, ingredientUnit string
		
		err := rows.Scan(
			&pi.ProductID,
			&pi.IngredientID,
			&pi.Quantity,
			&ingredientName,
			&ingredientUnit,
		)
		
		if err != nil {
			return nil, err
		}
		
		pi.IngredientName = ingredientName
		pi.IngredientUnit = ingredientUnit
		
		ingredients = append(ingredients, &pi)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return ingredients, nil
}

// CreateReservation creates a new inventory reservation for an order
func (r *InventoryRepository) CreateReservation(ctx context.Context, tx *sql.Tx, reservation *domain.InventoryReservation) error {
	query := `
		INSERT INTO inventory_reservations 
		(order_id, ingredient_id, quantity, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`
	
	now := time.Now()
	
	_, err := tx.ExecContext(
		ctx,
		query,
		reservation.OrderID,
		reservation.IngredientID,
		reservation.Quantity,
		reservation.Status,
		now,
	)
	
	return err
}

// UpdateReservationStatus updates the status of an inventory reservation
func (r *InventoryRepository) UpdateReservationStatus(ctx context.Context, tx *sql.Tx, orderID uint, status string) error {
	query := `
		UPDATE inventory_reservations
		SET status = $1, updated_at = $2
		WHERE order_id = $3
	`
	
	_, err := tx.ExecContext(ctx, query, status, time.Now(), orderID)
	return err
}

// GetReservationsByOrder retrieves all reservations for an order
func (r *InventoryRepository) GetReservationsByOrder(ctx context.Context, tx *sql.Tx, orderID uint) ([]*domain.InventoryReservation, error) {
	query := `
		SELECT id, order_id, ingredient_id, quantity, status, created_at, updated_at
		FROM inventory_reservations
		WHERE order_id = $1 AND status = 'reserved'
		FOR UPDATE
	`
	
	rows, err := tx.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var reservations []*domain.InventoryReservation
	for rows.Next() {
		var res domain.InventoryReservation
		var createdAt, updatedAt time.Time
		
		err := rows.Scan(
			&res.ID,
			&res.OrderID,
			&res.IngredientID,
			&res.Quantity,
			&res.Status,
			&createdAt,
			&updatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		res.CreatedAt = createdAt
		res.UpdatedAt = updatedAt
		
		reservations = append(reservations, &res)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return reservations, nil
}

// UpdateInventoryQuantity updates the quantity of an inventory item
func (r *InventoryRepository) UpdateInventoryQuantity(ctx context.Context, tx *sql.Tx, ingredientID uint, quantityChange float64) error {
	query := `
		UPDATE inventory_items
		SET current_quantity = current_quantity + $1, updated_at = $2
		WHERE id = $3
	`
	
	_, err := tx.ExecContext(ctx, query, quantityChange, time.Now(), ingredientID)
	return err
}

// CompleteOrderInventory marks all reservations for an order as completed
func (r *InventoryRepository) CompleteOrderInventory(ctx context.Context, tx *sql.Tx, orderID uint) error {
	query := `
		UPDATE inventory_reservations
		SET status = 'completed', updated_at = $1
		WHERE order_id = $2 AND status = 'reserved'
	`
	
	_, err := tx.ExecContext(ctx, query, time.Now(), orderID)
	return err
}

// CancelOrderInventory returns reserved inventory quantities and marks reservations as canceled
func (r *InventoryRepository) CancelOrderInventory(ctx context.Context, tx *sql.Tx, orderID uint) error {
	// First get all reserved quantities
	reservations, err := r.GetReservationsByOrder(ctx, tx, orderID)
	if err != nil {
		return err
	}
	
	// Return quantities to inventory
	for _, res := range reservations {
		err = r.UpdateInventoryQuantity(ctx, tx, res.IngredientID, res.Quantity)
		if err != nil {
			return err
		}
	}
	
	// Mark reservations as canceled
	query := `
		UPDATE inventory_reservations
		SET status = 'canceled', updated_at = $1
		WHERE order_id = $2 AND status = 'reserved'
	`
	
	_, err = tx.ExecContext(ctx, query, time.Now(), orderID)
	return err
}

// CheckProductAvailability verifies if a product has sufficient ingredients available
func (r *InventoryRepository) CheckProductAvailability(ctx context.Context, productID uint) (bool, error) {
	// Get all ingredients required for this product
	ingredients, err := r.GetByProductID(ctx, productID)
	fmt.Println("Ingredients: ", ingredients)
	if err != nil {
		fmt.Println("Error: ", err)
		return false, err
	}
	
	// If product has no ingredients, consider it available
	if len(ingredients) == 0 {
		return true, nil
	}
	
	// Check if all ingredients are available in sufficient quantity
	for _, ingredient := range ingredients {
		query := `
			SELECT current_quantity 
			FROM inventory_items 
			WHERE id = $1
		`
		
		var currentQuantity float64
		err := r.db.QueryRowContext(ctx, query, ingredient.IngredientID).Scan(&currentQuantity)
		if err != nil {
			return false, err
		}
		fmt.Println("Current Quantity: ", currentQuantity)
		fmt.Println("Required Quantity: ", ingredient.Quantity)
		if currentQuantity < ingredient.Quantity {
			return false, nil
		}
	}
	
	return true, nil
}

// CheckProductsAvailability checks availability for multiple products
func (r *InventoryRepository) CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error) {
	availability := make(map[uint]bool)
	
	for _, id := range productIDs {
		fmt.Println("Checking availability for product: ", id)
		available, err := r.CheckProductAvailability(ctx, id)
		fmt.Println("Available: ", available)
		if err != nil {
			return nil, err
		}
		availability[id] = available
	}
	fmt.Println("Availability: ", availability)
	
	return availability, nil
} 