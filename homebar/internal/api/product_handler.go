package api

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/service"
)

type ProductHandler struct {
	productService    *service.ProductService
	ingredientService IngredientServiceInterface
}

func NewProductHandler(ps *service.ProductService, is IngredientServiceInterface) *ProductHandler {
	return &ProductHandler{productService: ps, ingredientService: is}
}

// Create POST /api/products
func (h *ProductHandler) Create(c *gin.Context) {
	// limit the size of product image to be smaller than 10 mb
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form"})
		return
	}

	name := c.PostForm("name")
	priceStr := c.PostForm("price")
	category := c.PostForm("category")
	merchantID, _ := strconv.Atoi(c.PostForm("merchant_id"))
	isAvail := c.PostForm("is_available") == "true"
	description := c.PostForm("description")

	price, _ := strconv.ParseFloat(priceStr, 64)

	file, hdr, err := c.Request.FormFile("image")
	var mime string
	var data []byte
	if err == nil {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
			}
		}(file)
		mime = hdr.Header.Get("Content-Type")
		buf, _ := io.ReadAll(file)
		data = buf
	}

	product := &domain.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		MerchantID:  uint(merchantID),
		IsAvailable: isAvail,
		MimeType:    mime,
		ImageData:   data,
	}

	created, err := h.productService.Create(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetByID GET /api/products/:id
func (h *ProductHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		fmt.Println("Error parsing product id:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.productService.GetByID(c.Request.Context(), uint(id64))
	if err != nil {
		fmt.Println("Error getting product by id:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

// Update PUT /api/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("Error parsing product id:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form"})
		return
	}

	price, err := strconv.ParseFloat(c.PostForm("price"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price"})
		return
	}

	merchantID64, err := strconv.ParseUint(c.PostForm("merchant_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
		return
	}

	var mime string
	var data []byte
	file, hdr, err := c.Request.FormFile("image")
	if err == nil {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
			}
		}(file)

		mime = hdr.Header.Get("Content-Type")
		if mime != "image/jpeg" && mime != "image/png" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only jpeg/png allowed"})
			return
		}
		buf, _ := io.ReadAll(file)
		data = buf
	} else if !errors.Is(err, http.ErrMissingFile) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := &domain.Product{
		ID:          uint(id64),
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Price:       price,
		Category:    c.PostForm("category"),
		MerchantID:  uint(merchantID64),
		IsAvailable: c.PostForm("is_available") == "true",
	}

	if len(data) > 0 {
		product.MimeType = mime
		product.ImageData = data
	}

	updated, err := h.productService.Update(c, product)
	if err != nil {
		fmt.Println("Error updating product:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete DELETE /api/products/:id
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

// GetByMerchant GET /api/products/merchant/:id
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

// GetAll GET /api/products
func (h *ProductHandler) GetAll(c *gin.Context) {
	products, err := h.productService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// GetImage GET /api/products/:id/image
func (h *ProductHandler) GetImage(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.productService.GetByID(c, uint(id64))
	if err != nil || len(p.ImageData) == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Data(http.StatusOK, p.MimeType, p.ImageData)
}

// CheckAvailability handles GET /api/products/availability
func (h *ProductHandler) CheckAvailability(c *gin.Context) {
	var req struct {
		ProductIDs []uint `json:"product_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	availability, err := h.ingredientService.CheckProductsAvailability(c.Request.Context(), req.ProductIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check product availability"})
		return
	}

	fmt.Println("Availability: ", availability)

	c.JSON(http.StatusOK, gin.H{"availability": availability})
}
