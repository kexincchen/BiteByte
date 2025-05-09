package domain

import (
	"time"
)

// InventoryReservation represents a temporary lock on inventory during an order
type InventoryReservation struct {
	ID           uint      `json:"id"`
	OrderID      uint      `json:"order_id"`
	IngredientID uint      `json:"ingredient_id"`
	Quantity     float64   `json:"quantity"`
	Status       string    `json:"status"` // reserved, completed, canceled
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
