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
