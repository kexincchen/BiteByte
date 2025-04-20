package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string          `json:"username" binding:"required"`
		Email    string          `json:"email" binding:"required,email"`
		Password string          `json:"password" binding:"required,min=6"`
		Role     domain.UserRole `json:"role" binding:"required"`
		// Customer specific fields
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Address   string `json:"address"`
		Phone     string `json:"phone"`
		// Merchant specific fields
		BusinessName string `json:"business_name"`
		Description  string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role-specific required fields
	if req.Role == domain.RoleCustomer {
		if req.FirstName == "" || req.LastName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "first_name and last_name are required for customers"})
			return
		}
	} else if req.Role == domain.RoleMerchant {
		if req.BusinessName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "business_name is required for merchants"})
			return
		}
	}

	// Create the appropriate user type based on role
	var result interface{}
	var err error

	if req.Role == domain.RoleCustomer {
		customer := &domain.Customer{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Address:   req.Address,
			Phone:     req.Phone,
		}
		result, err = h.userService.RegisterCustomer(
			c.Request.Context(),
			req.Username,
			req.Email,
			req.Password,
			customer,
		)
	} else if req.Role == domain.RoleMerchant {
		merchant := &domain.Merchant{
			BusinessName: req.BusinessName,
			Description:  req.Description,
			Address:      req.Address,
			Phone:        req.Phone,
			Username:     req.Username, // This seems redundant but matches the domain model
			IsVerified:   false,        // Default to unverified
		}
		result, err = h.userService.RegisterMerchant(
			c.Request.Context(),
			req.Username,
			req.Email,
			req.Password,
			merchant,
		)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// In a real application, you would generate a JWT token here
	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": "sample-jwt-token", // This would be a real JWT token
	})
}
