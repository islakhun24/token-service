package handler

import (
	"net/http"
	_ "strings"

	"token-service/internal/domain"
	"token-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RegistryHandler struct {
	service *service.RegistryService
	logger  *zap.Logger
}

func NewRegistryHandler(s *service.RegistryService, logger *zap.Logger) *RegistryHandler {
	return &RegistryHandler{service: s, logger: logger}
}

func (h *RegistryHandler) ListTokens(c *gin.Context) {
	search := c.Query("search")
	tokens, err := h.service.ListTokens(c.Request.Context(), search)
	if err != nil {
		h.logger.Error("Failed to list tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": tokens, "total": len(tokens)})
}

func (h *RegistryHandler) GetToken(c *gin.Context) {
	canonical := c.Param("canonical")
	token, err := h.service.GetToken(c.Request.Context(), canonical)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		return
	}
	c.JSON(http.StatusOK, token)
}

func (h *RegistryHandler) CreateToken(c *gin.Context) {
	var req domain.CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.CreateToken(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create token", zap.Error(err))
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, token)
}

func (h *RegistryHandler) UpdateToken(c *gin.Context) {
	canonical := c.Param("canonical")
	var req domain.CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.UpdateToken(c.Request.Context(), canonical, req)
	if err != nil {
		h.logger.Error("Failed to update token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *RegistryHandler) DeleteToken(c *gin.Context) {
	canonical := c.Param("canonical")
	if err := h.service.DeleteToken(c.Request.Context(), canonical); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *RegistryHandler) VerifyArkham(c *gin.Context) {
	canonical := c.Param("canonical")
	token, err := h.service.VerifyArkham(c.Request.Context(), canonical)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, token)
}

func (h *RegistryHandler) ToggleActive(c *gin.Context) {
	canonical := c.Param("canonical")
	token, err := h.service.ToggleActive(c.Request.Context(), canonical)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, token)
}

func (h *RegistryHandler) GetExchangeSymbols(c *gin.Context) {
	exchange := c.Param("exchange")
	canonical := c.Query("canonical")

	symbols, err := h.service.GetExchangeSymbols(c.Request.Context(), canonical, exchange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": symbols})
}

func (h *RegistryHandler) Normalize(c *gin.Context) {
	exchange := c.Query("exchange")
	symbol := c.Query("symbol")

	if exchange == "" || symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exchange and symbol required"})
		return
	}

	result, err := h.service.Normalize(c.Request.Context(), exchange, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *RegistryHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
