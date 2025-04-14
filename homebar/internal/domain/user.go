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
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Customer struct {
	UserID    uint   `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

type Merchant struct {
	UserID       uint   `json:"user_id"`
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	IsVerified   bool   `json:"is_verified"`
} 