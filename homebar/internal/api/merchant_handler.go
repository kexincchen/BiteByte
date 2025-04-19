package api

import (
	"github.com/kexincchen/homebar/internal/domain"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kexincchen/homebar/internal/service"
)

type MerchantHandler struct{ svc *service.MerchantService }

func NewMerchantHandler(s *service.MerchantService) *MerchantHandler { return &MerchantHandler{s} }

func (h *MerchantHandler) List(c *gin.Context) {
	list, _ := h.svc.List(c)
	c.JSON(http.StatusOK, list)
}

func (h *MerchantHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.svc.GetByID(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *MerchantHandler) GetByUsername(c *gin.Context) {
	m, err := h.svc.GetByUsername(c, c.Param("username"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

// Create POST /api/merchants
func (h *MerchantHandler) Create(c *gin.Context) {
	var req struct {
		UserID       uint   `json:"user_id"       binding:"required"`
		BusinessName string `json:"business_name" binding:"required"`
		Description  string `json:"description"`
		Address      string `json:"address"`
		Phone        string `json:"phone"`
		Username     string `json:"username"      binding:"required"`
		IsVerified   bool   `json:"is_verified"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &domain.Merchant{
		UserID:       req.UserID,
		BusinessName: req.BusinessName,
		Description:  req.Description,
		Address:      req.Address,
		Phone:        req.Phone,
		Username:     req.Username,
		IsVerified:   req.IsVerified,
	}
	if err := h.svc.Create(c, m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

// Update PUT /api/merchants/:id
func (h *MerchantHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		BusinessName string `json:"business_name"`
		Description  string `json:"description"`
		Address      string `json:"address"`
		Phone        string `json:"phone"`
		Username     string `json:"username"`
		IsVerified   bool   `json:"is_verified"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m := &domain.Merchant{
		ID:           uint(id),
		BusinessName: req.BusinessName,
		Description:  req.Description,
		Address:      req.Address,
		Phone:        req.Phone,
		Username:     req.Username,
		IsVerified:   req.IsVerified,
		UpdatedAt:    time.Now(),
	}
	if err := h.svc.Update(c, m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

// Delete DELETE /api/merchants/:id
func (h *MerchantHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.Delete(c, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
