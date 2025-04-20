package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/service"
	"net/http"
	"strconv"
	"fmt"
)

type OrderHandler struct{ svc *service.OrderService }

func NewOrderHandler(s *service.OrderService) *OrderHandler { return &OrderHandler{svc: s} }

// Create POST /api/orders
func (h *OrderHandler) Create(c *gin.Context) {
	var req struct {
		CustomerID uint `json:"customer_id"`
		MerchantID uint `json:"merchant_id"`
		Items      []struct {
			ProductID uint    `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order"})
		return
	}

	var items []service.SimpleItem
	for _, it := range req.Items {
		items = append(items, service.SimpleItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Price:     it.Price,
		})
	}

	order, err := h.svc.CreateOrder(c, req.CustomerID, req.MerchantID, items, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// GetByID GET /api/orders/:id
func (h *OrderHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	o, items, err := h.svc.GetByID(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": o, "items": items})
}

// List GET /api/orders?customer=1  or  ?merchant=2
func (h *OrderHandler) List(c *gin.Context) {
	if cidStr := c.Query("customer"); cidStr != "" {
		cid, _ := strconv.Atoi(cidStr)
		list, _ := h.svc.ListByCustomer(c, uint(cid))
		c.JSON(http.StatusOK, list)
		return
	}
	if midStr := c.Query("merchant"); midStr != "" {
		mid, _ := strconv.Atoi(midStr)
		list, _ := h.svc.ListByMerchant(c, uint(mid))
		c.JSON(http.StatusOK, list)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "missing filter"})
}

// GetByUser retrieves orders for any user (merchant or customer) based on role
func (h *OrderHandler) GetByUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get the role from query parameter or from JWT token
	role := c.Query("role")
	if role == "" {
		// If not provided in query, try to get from the authenticated user
		// This would require middleware to set user info in the context
		if userInfo, exists := c.Get("user"); exists {
			if user, ok := userInfo.(map[string]interface{}); ok {
				role = user["role"].(string)
			}
		}
	}

	var orders interface{}
	var fetchErr error

	switch role {
	case "merchant":
		// Get merchant ID from user ID
		merchant, err := h.merchantService.GetByUserID(c.Request.Context(), uint(userID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("merchant not found for user ID %d", userID)})
			return
		}
		orders, fetchErr = h.svc.GetByMerchant(c.Request.Context(), merchant.ID)
	case "customer":
		// Get customer ID from user ID or use user ID directly if your system allows
		orders, fetchErr = h.svc.GetByCustomer(c.Request.Context(), uint(userID))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or missing role parameter"})
		return
	}

	if fetchErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fetchErr.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}
