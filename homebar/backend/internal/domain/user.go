package domain

import (
	"time"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleMerchant UserRole = "merchant"
)

type User struct {
	ID         uint      `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	Role       UserRole  `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	MerchantID int64     `json:"merchant_id,omitempty"`
}

type Customer struct {
	UserID    uint   `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

type Merchant struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	BusinessName string    `json:"business_name"`
	Description  string    `json:"description"`
	Address      string    `json:"address"`
	Phone        string    `json:"phone"`
	Username     string    `json:"username"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
