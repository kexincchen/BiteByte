package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/service"
	"net/http"
	"strconv"
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
