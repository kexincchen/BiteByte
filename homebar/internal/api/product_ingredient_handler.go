package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/service"
)

type ProductIngredientHandler struct {
	productIngredientService *service.ProductIngredientService
	productService           *service.ProductService
	ingredientService        *service.IngredientService
}

func NewProductIngredientHandler(
	productIngredientService *service.ProductIngredientService,
	productService *service.ProductService,
	ingredientService *service.IngredientService,
) *ProductIngredientHandler {
	return &ProductIngredientHandler{
		productIngredientService: productIngredientService,
		productService:           productService,
		ingredientService:        ingredientService,
	}
}

// GetByProductID gets all ingredients for a product
func (h *ProductIngredientHandler) GetByProductID(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Verify the product exists
	product, err := h.productService.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Get product ingredients
	ingredients, err := h.productIngredientService.GetProductIngredients(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product ingredients"})
		return
	}

	c.JSON(http.StatusOK, ingredients)
}

// GetByIngredientID gets all products that use a specific ingredient
func (h *ProductIngredientHandler) GetByIngredientID(c *gin.Context) {
	// This would require a different repo method that we haven't implemented yet
	// For now, we'll return a not implemented response
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Method not implemented"})
}

// Create adds an ingredient to a product
func (h *ProductIngredientHandler) Create(c *gin.Context) {
	var request struct {
		ProductID    int64   `json:"product_id"`
		IngredientID int64   `json:"ingredient_id"`
		Quantity     float64 `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify the product exists and authorize the merchant
	product, err := h.productService.GetProductByID(c.Request.Context(), uint(request.ProductID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Authorize merchant access
	if !h.authorizeMerchant(c, int64(product.MerchantID)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify the ingredient exists and belongs to the merchant
	ingredient, err := h.ingredientService.GetIngredientByID(c.Request.Context(), request.IngredientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving ingredient"})
		return
	}

	if ingredient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ingredient not found"})
		return
	}

	if ingredient.MerchantID != int64(product.MerchantID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ingredient does not belong to product's merchant"})
		return
	}

	// Add the ingredient to the product
	err = h.productIngredientService.AddIngredientToProduct(
		c.Request.Context(),
		request.ProductID,
		request.IngredientID,
		request.Quantity,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding ingredient to product"})
		return
	}

	// Create response with ingredient details
	response := struct {
		ProductID    int64   `json:"product_id"`
		IngredientID int64   `json:"ingredient_id"`
		Quantity     float64 `json:"quantity"`
		Name         string  `json:"name"`
		Unit         string  `json:"unit"`
	}{
		ProductID:    request.ProductID,
		IngredientID: ingredient.ID,
		Quantity:     request.Quantity,
		Name:         ingredient.Name,
		Unit:         ingredient.Unit,
	}

	c.JSON(http.StatusCreated, response)
}

// GetByID gets a specific product-ingredient relationship
func (h *ProductIngredientHandler) GetByID(c *gin.Context) {
	// This would require a different model structure with a composite primary key
	// For now, we'll return a not implemented response
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Method not implemented"})
}

// Update updates a product-ingredient relationship
func (h *ProductIngredientHandler) Update(c *gin.Context) {
	var request struct {
		ProductID    int64   `json:"product_id"`
		IngredientID int64   `json:"ingredient_id"`
		Quantity     float64 `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify the product exists and authorize the merchant
	product, err := h.productService.GetProductByID(c.Request.Context(), uint(request.ProductID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Authorize merchant access
	if !h.authorizeMerchant(c, int64(product.MerchantID)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// This is the same as Create since our AddIngredientToProduct handles the upsert
	err = h.productIngredientService.AddIngredientToProduct(
		c.Request.Context(),
		request.ProductID,
		request.IngredientID,
		request.Quantity,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating product ingredient"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product ingredient updated successfully"})
}

// Delete removes an ingredient from a product
func (h *ProductIngredientHandler) Delete(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	ingredientID, err := strconv.ParseInt(c.Param("ingredientId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredient ID"})
		return
	}

	// Verify the product exists and authorize the merchant
	product, err := h.productService.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Authorize merchant access
	if !h.authorizeMerchant(c, int64(product.MerchantID)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Remove the ingredient from the product
	err = h.productIngredientService.RemoveIngredientFromProduct(c.Request.Context(), productID, ingredientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing ingredient from product"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper method to authorize merchant access
func (h *ProductIngredientHandler) authorizeMerchant(c *gin.Context, merchantID int64) bool {
	// Get the user from context
	userRaw, exists := c.Get("user")
	if !exists {
		return false
	}

	user, ok := userRaw.(*domain.User)
	if !ok || user == nil {
		return false
	}

	// Check if user is a merchant and matches the merchant ID
	if user.Role != "merchant" || user.MerchantID != merchantID {
		return false
	}

	return true
}
