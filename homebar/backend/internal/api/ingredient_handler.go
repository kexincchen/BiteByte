package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
)

// Update the IngredientHandler struct to support both regular and Raft-enabled services
type IngredientHandler struct {
	service IngredientServiceInterface
}

// Define an interface that both IngredientService and RaftIngredientService implement
type IngredientServiceInterface interface {
	CreateIngredient(ctx context.Context, ingredient *domain.Ingredient) (*domain.Ingredient, error)
	GetIngredientByID(ctx context.Context, id int64) (*domain.Ingredient, error)
	GetIngredientsByMerchant(ctx context.Context, merchantID int64) ([]*domain.Ingredient, error)
	UpdateIngredient(ctx context.Context, ingredient *domain.Ingredient) error
	DeleteIngredient(ctx context.Context, id int64) error
	GetInventorySummary(ctx context.Context, merchantID int64) (map[string]interface{}, error)
	CheckProductAvailability(ctx context.Context, productID uint) (bool, error)
	CheckProductsAvailability(ctx context.Context, productIDs []uint) (map[uint]bool, error)
}

// NewIngredientHandler creates a new handler with either service type
func NewIngredientHandler(service IngredientServiceInterface) *IngredientHandler {
	return &IngredientHandler{service: service}
}

// Create handles ingredient creation
func (h *IngredientHandler) Create(c *gin.Context) {
	var ingredient domain.Ingredient
	if err := c.ShouldBindJSON(&ingredient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get merchant ID from URL
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	// Set merchant ID from path parameter
	ingredient.MerchantID = merchantID

	// Set timestamps
	now := time.Now()
	ingredient.CreatedAt = now
	ingredient.UpdatedAt = now

	createdIngredient, err := h.service.CreateIngredient(c.Request.Context(), &ingredient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating ingredient"})
		return
	}

	c.JSON(http.StatusCreated, createdIngredient)
}

// GetByID handles getting an ingredient by ID
func (h *IngredientHandler) GetByID(c *gin.Context) {
	// Get ingredient ID from URL
	ingredientID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredient ID"})
		return
	}

	// Get merchant ID from URL
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	ingredient, err := h.service.GetIngredientByID(c.Request.Context(), ingredientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving ingredient"})
		return
	}

	if ingredient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ingredient not found"})
		return
	}

	// Verify the ingredient belongs to the merchant
	if ingredient.MerchantID != merchantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Ingredient does not belong to merchant"})
		return
	}

	c.JSON(http.StatusOK, ingredient)
}

// Update handles updating an ingredient
func (h *IngredientHandler) Update(c *gin.Context) {
	// Get ingredient ID from URL
	ingredientID, err := strconv.ParseInt(c.Param("ingredientId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredient ID"})
		return
	}

	// Get merchant ID from URL
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	// Get the existing ingredient
	existingIngredient, err := h.service.GetIngredientByID(c.Request.Context(), ingredientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving ingredient"})
		return
	}

	if existingIngredient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ingredient not found"})
		return
	}

	// Verify the ingredient belongs to the merchant
	if existingIngredient.MerchantID != merchantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Ingredient does not belong to merchant"})
		return
	}

	// Parse update data
	var updateData domain.Ingredient
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update only allowed fields
	existingIngredient.Name = updateData.Name
	existingIngredient.Quantity = updateData.Quantity
	existingIngredient.Unit = updateData.Unit
	existingIngredient.LowStockThreshold = updateData.LowStockThreshold
	existingIngredient.UpdatedAt = time.Now()

	if err := h.service.UpdateIngredient(c.Request.Context(), existingIngredient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating ingredient"})
		return
	}

	c.JSON(http.StatusOK, existingIngredient)
}

// Delete handles deleting an ingredient
func (h *IngredientHandler) Delete(c *gin.Context) {
	// Get ingredient ID from URL
	ingredientID, err := strconv.ParseInt(c.Param("ingredientId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredient ID"})
		return
	}
	fmt.Println("Ingredient ID: ", ingredientID)

	// Get merchant ID from URL
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	// Get the existing ingredient to verify ownership
	existingIngredient, err := h.service.GetIngredientByID(c.Request.Context(), ingredientID)
	fmt.Println("Existing ingredient: ", existingIngredient)
	if err != nil {
		fmt.Println("Error retrieving ingredient: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving ingredient"})
		return
	}

	if existingIngredient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ingredient not found"})
		return
	}

	// Verify the ingredient belongs to the merchant
	if existingIngredient.MerchantID != merchantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Ingredient does not belong to merchant"})
		return
	}

	// Delete the ingredient
	if err := h.service.DeleteIngredient(c.Request.Context(), ingredientID); err != nil {
		fmt.Println("Error deleting ingredient: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting ingredient"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAll handles getting all ingredients for a merchant
func (h *IngredientHandler) GetAll(c *gin.Context) {
	// Get merchant ID from URL
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	fmt.Println("Merchant ID: ", merchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	ingredients, err := h.service.GetIngredientsByMerchant(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving ingredients"})
		return
	}

	c.JSON(http.StatusOK, ingredients)
}

// GetInventorySummary returns summary statistics for a merchant's inventory
func (h *IngredientHandler) GetInventorySummary(c *gin.Context) {
	// Get merchant ID from URL
	merchantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	summary, err := h.service.GetInventorySummary(c.Request.Context(), merchantID)
	if err != nil {
		fmt.Println("Error retrieving inventory summary: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving inventory summary"})
		return
	}

	fmt.Println("Inventory summary: ", summary)

	c.JSON(http.StatusOK, summary)
}

