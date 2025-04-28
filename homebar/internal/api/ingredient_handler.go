package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/service"
)

type IngredientHandler struct {
	ingredientService *service.IngredientService
}

func NewIngredientHandler(ingredientService *service.IngredientService) *IngredientHandler {
	return &IngredientHandler{
		ingredientService: ingredientService,
	}
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

	// Authorize merchant access
	// if !h.authorizeMerchant(c, merchantID) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	// Set merchant ID from path parameter
	ingredient.MerchantID = merchantID

	// Set timestamps
	now := time.Now()
	ingredient.CreatedAt = now
	ingredient.UpdatedAt = now

	createdIngredient, err := h.ingredientService.CreateIngredient(c.Request.Context(), &ingredient)
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

	// Authorize merchant access
	// if !h.authorizeMerchant(c, merchantID) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	ingredient, err := h.ingredientService.GetIngredientByID(c.Request.Context(), ingredientID)
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

	// Authorize merchant access
	// if !h.authorizeMerchant(c, merchantID) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	// Get the existing ingredient
	existingIngredient, err := h.ingredientService.GetIngredientByID(c.Request.Context(), ingredientID)
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

	if err := h.ingredientService.UpdateIngredient(c.Request.Context(), existingIngredient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating ingredient"})
		return
	}

	c.JSON(http.StatusOK, existingIngredient)
}

// Delete handles deleting an ingredient
func (h *IngredientHandler) Delete(c *gin.Context) {
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

	// Authorize merchant access
	// if !h.authorizeMerchant(c, merchantID) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	// Get the existing ingredient to verify ownership
	existingIngredient, err := h.ingredientService.GetIngredientByID(c.Request.Context(), ingredientID)
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

	// Delete the ingredient
	if err := h.ingredientService.DeleteIngredient(c.Request.Context(), ingredientID); err != nil {
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

	// Authorize merchant access
	// if !h.authorizeMerchant(c, merchantID) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	ingredients, err := h.ingredientService.GetIngredientsByMerchant(c.Request.Context(), merchantID)
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

	// Authorize merchant access
	// if !h.authorizeMerchant(c, merchantID) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	summary, err := h.ingredientService.GetInventorySummary(c.Request.Context(), merchantID)
	if err != nil {
		fmt.Println("Error retrieving inventory summary: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving inventory summary"})
		return
	}

	fmt.Println("Inventory summary: ", summary)

	c.JSON(http.StatusOK, summary)
}

// Helper method to authorize merchant access
func (h *IngredientHandler) authorizeMerchant(c *gin.Context, merchantID int64) bool {
	// Get the user from context
	userRaw, exists := c.Get("user")
	if !exists {
		fmt.Println("User not found in context")
		return false
	}

	user, ok := userRaw.(*domain.User)
	if !ok || user == nil {
		fmt.Println("User is not a domain.User")
		return false
	}

	// Check if user is a merchant and matches the merchant ID
	if user.Role != "merchant" || user.ID != uint(merchantID) {
		fmt.Println("User is not a merchant or does not match merchant ID")
		return false
	}

	return true
}
