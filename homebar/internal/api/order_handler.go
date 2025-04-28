package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/service"
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
		fmt.Println("cidStr: ", cidStr)
		cid, _ := strconv.Atoi(cidStr)
		list, _ := h.svc.ListByCustomer(c, uint(cid))
		fmt.Println("list: ", list)
		c.JSON(http.StatusOK, list)
		return
	}
	if midStr := c.Query("merchant"); midStr != "" {
		fmt.Println("merchant midStr: ", midStr)
		mid, _ := strconv.Atoi(midStr)
		list, _ := h.svc.ListByMerchant(c, uint(mid))
		c.JSON(http.StatusOK, list)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "missing filter"})
}

// func (s *OrderService) UpdateOrder(ctx context.Context, id uint, status string, notes string, deliveryAddr string, deliveryTime string) error {
// 	// Get the existing order
// 	order, _, err := s.orderRepo.GetByID(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	// Update fields that are provided
// 	if status != "" {
// 		order.Status = domain.OrderStatus(status)
// 	}

// 	if notes != "" {
// 		order.Notes = notes
// 	}

// 	if deliveryAddr != "" {
// 		order.DeliveryAddr = deliveryAddr
// 	}

// 	// Update the order in the database
// 	return s.orderRepo.UpdateOrder(ctx, order, deliveryTime)
// }

// UpdateStatus PUT /api/orders/:id/status
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Convert string to OrderStatus type
	status := domain.OrderStatus(req.Status)
	fmt.Println("status: ", status)
	// Validate status value
	validStatuses := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusPreparing,
		domain.OrderStatusReady,
		domain.OrderStatusDelivered,
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
	}

	isValid := false
	for _, s := range validStatuses {
		if status == s {
			isValid = true
			break
		}
	}

	if !isValid {
		fmt.Println("invalid status value")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	if err := h.svc.UpdateStatus(c, uint(id), status); err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("status updated successfully")

	c.Status(http.StatusOK)
}

// UpdateOrder PUT /api/orders/:id
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	fmt.Println("UpdateOrder called")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var orderUpdate struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&orderUpdate); err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate status if provided
	if orderUpdate.Status != "" {
		status := domain.OrderStatus(orderUpdate.Status)
		validStatuses := []domain.OrderStatus{
			domain.OrderStatusPending,
			domain.OrderStatusConfirmed,
			domain.OrderStatusPreparing,
			domain.OrderStatusReady,
			domain.OrderStatusDelivered,
			domain.OrderStatusCancelled,
			domain.OrderStatusRefunded,
		}

		isValid := false
		for _, s := range validStatuses {
			if status == s {
				isValid = true
				break
			}
		}

		if !isValid {
			fmt.Println("invalid status value")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
			return
		}
	}

	if err := h.svc.UpdateOrder(c, uint(id), orderUpdate.Status, orderUpdate.Notes); err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("order updated successfully")
	c.Status(http.StatusOK)
}
