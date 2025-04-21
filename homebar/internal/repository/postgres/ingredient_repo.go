package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
)

type IngredientRepository struct {
	db *sql.DB
}

func NewIngredientRepository(db *sql.DB) *IngredientRepository {
	return &IngredientRepository{
		db: db,
	}
}

// Create adds a new ingredient to the database
func (r *IngredientRepository) Create(ctx context.Context, ingredient *domain.Ingredient) (*domain.Ingredient, error) {
	query := `
		INSERT INTO ingredients (merchant_id, name, quantity, unit, low_stock_threshold, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	
	now := time.Now()
	ingredient.CreatedAt = now
	ingredient.UpdatedAt = now
	
	err := r.db.QueryRowContext(
		ctx,
		query,
		ingredient.MerchantID,
		ingredient.Name,
		ingredient.Quantity,
		ingredient.Unit,
		ingredient.LowStockThreshold,
		ingredient.CreatedAt,
		ingredient.UpdatedAt,
	).Scan(&ingredient.ID)
	
	if err != nil {
		return nil, err
	}
	
	return ingredient, nil
}

// GetByID retrieves an ingredient by its ID
func (r *IngredientRepository) GetByID(ctx context.Context, id int64) (*domain.Ingredient, error) {
	query := `
		SELECT id, merchant_id, name, quantity, unit, low_stock_threshold, created_at, updated_at
		FROM ingredients
		WHERE id = $1
	`
	
	var ingredient domain.Ingredient
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ingredient.ID,
		&ingredient.MerchantID,
		&ingredient.Name,
		&ingredient.Quantity,
		&ingredient.Unit,
		&ingredient.LowStockThreshold,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	
	return &ingredient, nil
}

// GetByMerchant retrieves all ingredients for a merchant
func (r *IngredientRepository) GetByMerchant(ctx context.Context, merchantID int64) ([]*domain.Ingredient, error) {
	query := `
		SELECT id, merchant_id, name, quantity, unit, low_stock_threshold, created_at, updated_at
		FROM ingredients
		WHERE merchant_id = $1
		ORDER BY name
	`
	
	rows, err := r.db.QueryContext(ctx, query, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ingredients []*domain.Ingredient
	for rows.Next() {
		var ingredient domain.Ingredient
		err := rows.Scan(
			&ingredient.ID,
			&ingredient.MerchantID,
			&ingredient.Name,
			&ingredient.Quantity,
			&ingredient.Unit,
			&ingredient.LowStockThreshold,
			&ingredient.CreatedAt,
			&ingredient.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, &ingredient)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return ingredients, nil
}

// Update updates an existing ingredient
func (r *IngredientRepository) Update(ctx context.Context, ingredient *domain.Ingredient) error {
	query := `
		UPDATE ingredients
		SET name = $1, quantity = $2, unit = $3, low_stock_threshold = $4, updated_at = $5
		WHERE id = $6
	`
	
	ingredient.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(
		ctx,
		query,
		ingredient.Name,
		ingredient.Quantity,
		ingredient.Unit,
		ingredient.LowStockThreshold,
		ingredient.UpdatedAt,
		ingredient.ID,
	)
	
	return err
}

// Delete removes an ingredient
func (r *IngredientRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM ingredients WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetInventorySummary gets summary statistics for a merchant's inventory
func (r *IngredientRepository) GetInventorySummary(ctx context.Context, merchantID int64) (map[string]interface{}, error) {
	totalQuery := `SELECT COUNT(*) FROM ingredients WHERE merchant_id = $1`
	lowStockQuery := `
		SELECT COUNT(*) 
		FROM ingredients 
		WHERE merchant_id = $1 AND quantity <= low_stock_threshold
	`
	
	var totalIngredients int
	var lowStockCount int
	
	err := r.db.QueryRowContext(ctx, totalQuery, merchantID).Scan(&totalIngredients)
	if err != nil {
		return nil, err
	}
	
	err = r.db.QueryRowContext(ctx, lowStockQuery, merchantID).Scan(&lowStockCount)
	if err != nil {
		return nil, err
	}

	fmt.Println("Total ingredients: ", totalIngredients)
	fmt.Println("Low stock count: ", lowStockCount)
	
	return map[string]interface{}{
		"totalIngredients": totalIngredients,
		"lowStockCount":    lowStockCount,
	}, nil
}

// LockInventoryForOrder attempts to lock inventory for an order
// Returns false if there's not enough inventory
func (r *IngredientRepository) LockInventoryForOrder(ctx context.Context, orderItems []*domain.OrderItem) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	
	// For each order item, get its product ingredients and check inventory
	for _, item := range orderItems {
		ingredients, err := r.getProductIngredients(ctx, tx, int64(item.ProductID))
		if err != nil {
			return false, err
		}
		
		// Check and update each ingredient
		for _, prodIngredient := range ingredients {
			ingredient, err := r.getIngredientWithLock(ctx, tx, prodIngredient.IngredientID)
			if err != nil {
				return false, err
			}
			
			// Calculate required quantity
			requiredQty := prodIngredient.Quantity * float64(item.Quantity)
			
			// Check if we have enough
			if ingredient.Quantity < requiredQty {
				return false, nil // Not enough inventory
			}
			
			// Update the ingredient quantity
			_, err = tx.ExecContext(
				ctx,
				"UPDATE ingredients SET quantity = quantity - $1 WHERE id = $2",
				requiredQty,
				ingredient.ID,
			)
			if err != nil {
				return false, err
			}
		}
	}
	
	// If we get here, everything is successful
	if err := tx.Commit(); err != nil {
		return false, err
	}
	
	return true, nil
}

// Helper to get a single ingredient with FOR UPDATE lock
func (r *IngredientRepository) getIngredientWithLock(ctx context.Context, tx *sql.Tx, id int64) (*domain.Ingredient, error) {
	query := `
		SELECT id, merchant_id, name, quantity, unit, low_stock_threshold, created_at, updated_at
		FROM ingredients
		WHERE id = $1
		FOR UPDATE
	`
	
	var ingredient domain.Ingredient
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&ingredient.ID,
		&ingredient.MerchantID,
		&ingredient.Name,
		&ingredient.Quantity,
		&ingredient.Unit,
		&ingredient.LowStockThreshold,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &ingredient, nil
}

// Helper to get ingredients for a product
func (r *IngredientRepository) getProductIngredients(ctx context.Context, tx *sql.Tx, productID int64) ([]*domain.ProductIngredient, error) {
	query := `
		SELECT product_id, ingredient_id, quantity
		FROM product_ingredients
		WHERE product_id = $1
	`
	
	rows, err := tx.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ingredients []*domain.ProductIngredient
	for rows.Next() {
		var ing domain.ProductIngredient
		err := rows.Scan(
			&ing.ProductID,
			&ing.IngredientID,
			&ing.Quantity,
		)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, &ing)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return ingredients, nil
} 