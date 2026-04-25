package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"futures-symbol-module/internal/symbol"
)

// FuturesHandler handles futures pairs API endpoints.
type FuturesHandler struct {
	service *symbol.Service
}

// NewFuturesHandler creates a new futures handler.
func NewFuturesHandler(service *symbol.Service) *FuturesHandler {
	return &FuturesHandler{service: service}
}

// GetPairs handles GET /futures/pairs.
func (h *FuturesHandler) GetPairs(c *gin.Context) {
	resp, err := h.service.GetFuturesPairs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
