package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/lib/pq"
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
		fmt.Println("Error: ", err)
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

// func (r *IngredientRepository) GetInventorySummary(ctx context.Context, merchantID int64) (map[string]interface{}, error) {
// 	summary := make(map[string]interface{})
	
// 	// Get total count of ingredients
// 	var totalCount int
// 	err := r.db.QueryRowContext(
// 		ctx,
// 		"SELECT COUNT(*) FROM ingredients WHERE merchant_id = $1",
// 		merchantID,
// 	).Scan(&totalCount)
	
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	summary["totalIngredients"] = totalCount
	
// 	// Get count of low stock items
// 	var lowStockCount int
// 	err = r.db.QueryRowContext(
// 		ctx,
// 		"SELECT COUNT(*) FROM ingredients WHERE merchant_id = $1 AND quantity <= low_stock_threshold",
// 		merchantID,
// 	).Scan(&lowStockCount)
	
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	summary["lowStockCount"] = lowStockCount
	
// 	return summary, nil
// }

// CheckProductsAvailability checks if multiple products have sufficient ingredients
func (r *IngredientRepository) CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	
	// Get product ingredients for all products in a single query
	query := `
		SELECT pi.product_id, pi.ingredient_id, pi.quantity, i.quantity as available_quantity
		FROM product_ingredients pi
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE pi.product_id = ANY($1)
	`
	
	// Convert []uint to []int64 for PostgreSQL ANY operator
	ids := make([]int64, len(productIDs))
	for i, id := range productIDs {
		ids[i] = int64(id)
		result[id] = true // Initialize all products as available
	}
	
	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	// Track which products we've checked to handle products with no ingredients
	productChecked := make(map[uint]bool)
	
	for rows.Next() {
		var productID, ingredientID int64
		var requiredQty, availableQty float64
		
		if err := rows.Scan(&productID, &ingredientID, &requiredQty, &availableQty); err != nil {
			return nil, err
		}
		
		productChecked[uint(productID)] = true
		
		// If there's not enough of this ingredient, mark the product as unavailable
		if availableQty < requiredQty {
			result[uint(productID)] = false
		}
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return result, nil
}

// LockInventoryForOrder attempts to lock inventory for an order
// Returns false if there's not enough inventory
func (r *IngredientRepository) LockInventoryForOrder(ctx context.Context, orderItems []*domain.OrderItem) (bool, error) {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Map to store required ingredients and their quantities
	requiredIngredients := make(map[int64]float64)

	// First pass: collect all required ingredients for all products
	for _, item := range orderItems {
		// Get the product ingredients
		query := `
			SELECT ingredient_id, quantity
			FROM product_ingredients
			WHERE product_id = $1
		`
		rows, err := tx.QueryContext(ctx, query, item.ProductID)
		if err != nil {
			return false, err
		}

		for rows.Next() {
			var ingredientID int64
			var quantity float64
			if err := rows.Scan(&ingredientID, &quantity); err != nil {
				rows.Close()
				return false, err
			}
			
			// Multiply by the order item quantity and add to our requirements
			requiredIngredients[ingredientID] += quantity * float64(item.Quantity)
		}
		rows.Close()
		
		if err = rows.Err(); err != nil {
			return false, err
		}
	}

	// Second pass: check and update each ingredient with locking
	for ingredientID, requiredQty := range requiredIngredients {
		// Lock the row for update
		var currentQty float64
		query := `
			SELECT quantity
			FROM ingredients
			WHERE id = $1
			FOR UPDATE
		`
		err := tx.QueryRowContext(ctx, query, ingredientID).Scan(&currentQty)
		if err != nil {
			return false, err
		}

		// Check if there's enough
		if currentQty < requiredQty {
			return false, nil // Not enough inventory
		}

		// Update the inventory
		_, err = tx.ExecContext(
			ctx,
			`UPDATE ingredients SET quantity = quantity - $1, updated_at = NOW() WHERE id = $2`,
			requiredQty,
			ingredientID,
		)
		if err != nil {
			return false, err
		}
	}

	// Everything is good, commit the transaction
	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

// RestoreInventoryForOrder restores ingredients that were previously locked for an order
func (r *IngredientRepository) RestoreInventoryForOrder(ctx context.Context, orderID uint) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First get the order items
	query := `
		SELECT product_id, quantity
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := tx.QueryContext(ctx, query, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var orderItems []*domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return err
		}
		orderItems = append(orderItems, &item)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// Map to store ingredients to be restored
	ingredientsToRestore := make(map[int64]float64)

	// Collect all ingredients used in this order
	for _, item := range orderItems {
		// Get the product ingredients
		query := `
			SELECT ingredient_id, quantity
			FROM product_ingredients
			WHERE product_id = $1
		`
		pRows, err := tx.QueryContext(ctx, query, item.ProductID)
		if err != nil {
			return err
		}

		for pRows.Next() {
			var ingredientID int64
			var quantity float64
			if err := pRows.Scan(&ingredientID, &quantity); err != nil {
				pRows.Close()
				return err
			}
			
			// Multiply by the order item quantity and add to our restoration map
			fmt.Println("ingredientID: ", ingredientID)
			fmt.Println("quantity: ", quantity)
			fmt.Println("item.Quantity: ", item.Quantity)
			ingredientsToRestore[ingredientID] += quantity * float64(item.Quantity)
		}
		pRows.Close()
		
		if err = pRows.Err(); err != nil {
			return err
		}
	}

	// Restore each ingredient
	for ingredientID, restoreQty := range ingredientsToRestore {
		_, err = tx.ExecContext(
			ctx,
			`UPDATE ingredients SET quantity = quantity + $1, updated_at = NOW() WHERE id = $2`,
			restoreQty,
			ingredientID,
		)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
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

// Add this method to the IngredientRepository
func (r *IngredientRepository) GetDB() *sql.DB {
	return r.db
}
