package api

import (
	"net/http"
	"strconv"

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
