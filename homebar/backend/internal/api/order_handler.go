package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/service"
)

type OrderHandler struct {
	orderService   service.OrderServiceInterface
	productService *service.ProductService
}

func NewOrderHandler(s service.OrderServiceInterface, ps *service.ProductService) *OrderHandler {
	return &OrderHandler{
		orderService:   s,
		productService: ps,
	}
}

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

	order, err := h.orderService.CreateOrder(c, req.CustomerID, req.MerchantID, items, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// GetByID GET /api/orders/:id
func (h *OrderHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	o, items, err := h.orderService.GetByID(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get each order item's product information
	itemsWithProducts := make([]map[string]interface{}, 0)
	for _, item := range items {
		product, err := h.productService.GetByID(c, item.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product details"})
			return
		}

		itemWithProduct := map[string]interface{}{
			"id":                  item.ID,
			"order_id":            item.OrderID,
			"product_id":          item.ProductID,
			"quantity":            item.Quantity,
			"price":               item.Price,
			"product_name":        product.Name,
			"product_description": product.Description,
		}
		itemsWithProducts = append(itemsWithProducts, itemWithProduct)
	}

	c.JSON(http.StatusOK, gin.H{
		"order": o,
		"items": itemsWithProducts,
	})
}

// List GET /api/orders?customer=1  or  ?merchant=2
func (h *OrderHandler) List(c *gin.Context) {
	if cidStr := c.Query("customer"); cidStr != "" {
		cid, _ := strconv.Atoi(cidStr)
		list, _ := h.orderService.ListByCustomer(c, uint(cid))
		c.JSON(http.StatusOK, list)
		return
	}
	if midStr := c.Query("merchant"); midStr != "" {
		fmt.Println("merchant midStr: ", midStr)
		mid, _ := strconv.Atoi(midStr)
		list, _ := h.orderService.ListByMerchant(c, uint(mid))
		c.JSON(http.StatusOK, list)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "missing filter"})
}


// UpdateStatus PUT /api/orders/:id/status
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate status value
	status := domain.OrderStatus(req.Status)
	validStatuses := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusCompleted,
		domain.OrderStatusCancelled,
	}

	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	// Update the status (this will also handle inventory)
	if err := h.orderService.UpdateStatus(c, uint(id), status); err != nil {
		if err.Error() == "invalid status transition" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// UpdateOrder PUT /api/orders/:id
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var orderUpdate struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&orderUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate status if provided
	if orderUpdate.Status != "" {
		status := domain.OrderStatus(orderUpdate.Status)
		validStatuses := []domain.OrderStatus{
			domain.OrderStatusPending,
			domain.OrderStatusCompleted,
			domain.OrderStatusCancelled,
		}

		isValid := false
		for _, s := range validStatuses {
			if status == s {
				isValid = true
				break
			}
		}

		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
			return
		}
	}

	if err := h.orderService.UpdateOrder(c, uint(id), orderUpdate.Status, orderUpdate.Notes); err != nil {
		if err.Error() == "invalid status transition" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// Add a new handler for checking product availability
func (h *OrderHandler) GetProductsAvailability(c *gin.Context) {
	var req struct {
		ProductIDs []uint `json:"product_ids"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	availability, err := h.orderService.CheckProductsAvailability(c, req.ProductIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check product availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"availability": availability})
}

// Delete DELETE /api/orders/:id
func (h *OrderHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	// Get the order first to check if it exists
	_, _, err = h.orderService.GetByID(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Delete the order
	if err := h.orderService.DeleteOrder(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
