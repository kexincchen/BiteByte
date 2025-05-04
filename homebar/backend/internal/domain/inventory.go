package domain

import (
	"time"
)

type Inventory struct {
	ID           uint      `json:"id"`
	MerchantID   uint      `json:"merchant_id"`
	IngredientID uint      `json:"ingredient_id"`
	Quantity     float64   `json:"quantity"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InventoryTransaction struct {
	ID           uint      `json:"id"`
	InventoryID  uint      `json:"inventory_id"`
	OrderID      *uint     `json:"order_id"`
	Quantity     float64   `json:"quantity"`
	Reason       string    `json:"reason"`
	PerformedBy  uint      `json:"performed_by"`
	TransactedAt time.Time `json:"transacted_at"`
}

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

// Product availability statuses
const (
	ProductAvailable = "available"
	ProductSoldOut   = "sold_out"
)
