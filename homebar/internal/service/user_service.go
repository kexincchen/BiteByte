package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo     repository.UserRepository
	customerRepo repository.CustomerRepository
	merchantRepo repository.MerchantRepository
	db           *sql.DB // For transactions
}

func NewUserService(
	userRepo repository.UserRepository,
	customerRepo repository.CustomerRepository,
	merchantRepo repository.MerchantRepository,
	db *sql.DB,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		merchantRepo: merchantRepo,
		db:           db,
	}
}

func (s *UserService) Register(ctx context.Context, username, email, password string, role domain.UserRole) (*domain.User, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	now := time.Now()
	user := &domain.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

// RegisterCustomer creates a user with customer details
func (s *UserService) RegisterCustomer(ctx context.Context, username, email, password string, customer *domain.Customer) (*domain.Customer, error) {
	// This would ideally be a transaction
	user, err := s.Register(ctx, username, email, password, domain.RoleCustomer)
	if err != nil {
		return nil, err
	}

	// Set the UserID from the newly created user
	customer.UserID = user.ID

	// Create the customer record - we'd need a customerRepo for this
	if err := s.customerRepo.Create(ctx, customer); err != nil {
		// In a transaction, we would rollback here
		// For now, we'll leave the user record as is
		return nil, err
	}

	return customer, nil
}

// RegisterMerchant creates a user with merchant details
func (s *UserService) RegisterMerchant(ctx context.Context, username, email, password string, merchant *domain.Merchant) (*domain.Merchant, error) {
	// This would ideally be a transaction
	user, err := s.Register(ctx, username, email, password, domain.RoleMerchant)
	if err != nil {
		return nil, err
	}

	// Set the UserID and timestamps from the newly created user
	merchant.UserID = user.ID
	merchant.CreatedAt = user.CreatedAt
	merchant.UpdatedAt = user.UpdatedAt

	// Create the merchant record - we'd need a merchantRepo for this
	if err := s.merchantRepo.Create(ctx, merchant); err != nil {
		// In a transaction, we would rollback here
		// For now, we'll leave the user record as is
		return nil, err
	}

	return merchant, nil
}
