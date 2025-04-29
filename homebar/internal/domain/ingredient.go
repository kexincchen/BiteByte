package domain

import (
	"time"
)

// Ingredient represents an ingredient used in products
type Ingredient struct {
	ID               int64     `json:"id"`
	MerchantID       int64     `json:"merchant_id"`
	Name             string    `json:"name"`
	Quantity         float64   `json:"quantity"`
	Unit             string    `json:"unit"`
	LowStockThreshold float64  `json:"low_stock_threshold"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Description      string    `json:"description"`
}

// ProductIngredient represents the relationship between a product and its ingredients
type ProductIngredient struct {
	ID             int64   `json:"id"`
	ProductID      int64   `json:"product_id"`
	IngredientID   int64   `json:"ingredient_id"`
	Quantity       float64 `json:"quantity"`
	IngredientName string  `json:"ingredient_name"`
	IngredientUnit string  `json:"ingredient_unit"`
}