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
	OrderID      *uint     `json:"order_id"` // Optional, can be null for manual adjustments
	Quantity     float64   `json:"quantity"` // Can be negative for deductions
	Reason       string    `json:"reason"`   // e.g., "order", "restock", "adjustment"
	PerformedBy  uint      `json:"performed_by"` // User ID
	TransactedAt time.Time `json:"transacted_at"`
} 