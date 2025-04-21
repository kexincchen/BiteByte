package postgres

import (
	"context"
	"database/sql"

	"github.com/kexincchen/homebar/internal/domain"
)

type ProductIngredientRepository struct {
	db *sql.DB
}

func NewProductIngredientRepository(db *sql.DB) *ProductIngredientRepository {
	return &ProductIngredientRepository{
		db: db,
	}
}

// AddIngredientToProduct links an ingredient to a product
func (r *ProductIngredientRepository) AddIngredientToProduct(ctx context.Context, productID, ingredientID int64, quantity float64) error {
	query := `
		INSERT INTO product_ingredients (product_id, ingredient_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (product_id, ingredient_id) 
		DO UPDATE SET quantity = $3
	`

	_, err := r.db.ExecContext(ctx, query, productID, ingredientID, quantity)
	return err
}


// RemoveIngredientFromProduct removes an ingredient from a product
func (r *ProductIngredientRepository) RemoveIngredientFromProduct(ctx context.Context, productID, ingredientID int64) error {
	query := `DELETE FROM product_ingredients WHERE product_id = $1 AND ingredient_id = $2`
	_, err := r.db.ExecContext(ctx, query, productID, ingredientID)
	return err
}

// GetProductIngredients gets all ingredients for a product with details
func (r *ProductIngredientRepository) GetProductIngredients(ctx context.Context, productID int64) ([]*domain.ProductIngredient, error) {
	query := `
		SELECT pi.product_id, pi.ingredient_id, pi.quantity, i.name, i.unit
		FROM product_ingredients pi
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE pi.product_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type extendedProductIngredient struct {
		domain.ProductIngredient
		Name string `json:"name"`
		Unit string `json:"unit"`
	}

	var ingredients []*domain.ProductIngredient
	for rows.Next() {
		var ing extendedProductIngredient
		err := rows.Scan(
			&ing.ProductID,
			&ing.IngredientID,
			&ing.Quantity,
			&ing.Name,
			&ing.Unit,
		)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, &domain.ProductIngredient{
			ProductID:    ing.ProductID,
			IngredientID: ing.IngredientID,
			Quantity:     ing.Quantity,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ingredients, nil
}
