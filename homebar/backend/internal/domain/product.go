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
	MimeType    string    `json:"-"`
	ImageData   []byte    `json:"-"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
