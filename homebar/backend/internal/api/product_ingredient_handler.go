package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/service"
)

type ProductIngredientHandler struct {
	productIngredientService *service.ProductIngredientService
	productService           *service.ProductService
	ingredientService        IngredientServiceInterface
}

func NewProductIngredientHandler(
	productIngredientService *service.ProductIngredientService,
	productService *service.ProductService,
	ingredientService IngredientServiceInterface,
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
		fmt.Println("Error getting product by id:", err)
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
		fmt.Println("Error getting product ingredients:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product ingredients"})
		return
	}

	c.JSON(http.StatusOK, ingredients)
}

// GetByIngredientID gets all products that use a specific ingredient
func (h *ProductIngredientHandler) GetByIngredientID(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Method not implemented"})
}

// Create adds an ingredient to a product
func (h *ProductIngredientHandler) Create(c *gin.Context) {
	fmt.Println("Create Product Ingredient...")
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
		fmt.Println("[Create] Error getting product by id:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
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
		fmt.Println("[Create] Error adding ingredient to product:", err)
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
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Method not implemented"})
}

// Update updates a product-ingredient relationship
func (h *ProductIngredientHandler) Update(c *gin.Context) {
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

	var request struct {
		Quantity float64 `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify the product exists and authorize the merchant
	product, err := h.productService.GetProductByID(c.Request.Context(), uint(productID))
	if err != nil {
		fmt.Println("[Update] Error getting product by id:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Update the ingredient quantity using the AddIngredientToProduct method
	err = h.productIngredientService.AddIngredientToProduct(
		c.Request.Context(),
		productID,
		ingredientID,
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
		fmt.Println("[Delete] Error getting product by id:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving product"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
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
