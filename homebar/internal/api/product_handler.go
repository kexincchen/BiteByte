package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/service"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(ps *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: ps}
}

// GET /api/products
func (h *ProductHandler) GetAll(c *gin.Context) {
	products, err := h.productService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// GET /api/products/:id
func (h *ProductHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.productService.GetByID(c.Request.Context(), uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

// GET /api/products/merchant/:id
func (h *ProductHandler) GetByMerchant(c *gin.Context) {
	merchantIDParam := c.Param("id")
	mid, err := strconv.ParseUint(merchantIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	products, err := h.productService.GetByMerchant(c.Request.Context(), uint(mid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// POST /api/products
func (h *ProductHandler) Create(c *gin.Context) {
	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		MerchantID  uint    `json:"merchant_id"`
		ImageURL    string  `json:"image_url"`
		IsAvailable bool    `json:"is_available"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product data"})
		return
	}

	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		MerchantID:  req.MerchantID,
		ImageURL:    req.ImageURL,
		IsAvailable: req.IsAvailable,
	}

	created, err := h.productService.Create(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// PUT /api/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		MerchantID  uint    `json:"merchant_id"`
		ImageURL    string  `json:"image_url"`
		IsAvailable bool    `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product data"})
		return
	}

	product := &domain.Product{
		ID:          uint(id64),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		MerchantID:  req.MerchantID,
		ImageURL:    req.ImageURL,
		IsAvailable: req.IsAvailable,
	}

	updated, err := h.productService.Update(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DELETE /api/products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.productService.Delete(c.Request.Context(), uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
