package domain

import (
	"time"
)

type Product struct {
	ID          uint      `json:"id"`
	MerchantID  uint      `json:"merchant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Ingredient struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
}

type ProductIngredient struct {
	ProductID    uint    `json:"product_id"`
	IngredientID uint    `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
}
