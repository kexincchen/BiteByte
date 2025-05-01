package domain

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// IsValidTransition checks if a status transition is allowed
func IsValidTransition(from, to OrderStatus) bool {
	// Only pending orders can transition to completed or cancelled
	if from == OrderStatusPending && (to == OrderStatusCompleted || to == OrderStatusCancelled) {
		return true
	}

	// Allow setting the same status (no change)
	if from == to {
		return true
	}

	// All other transitions are invalid
	return false
}

type Order struct {
	ID           uint        `json:"id"`
	CustomerID   uint        `json:"customer_id"`
	MerchantID   uint        `json:"merchant_id"`
	TotalAmount  float64     `json:"total_amount"`
	Status       OrderStatus `json:"status"`
	Notes        string      `json:"notes"`
	DeliveryAddr string      `json:"delivery_addr,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uint    `json:"id"`
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
