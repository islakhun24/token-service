package handler

import (
	"net/http"

	"token-service/internal/domain"
	"token-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SyncHandler struct {
	service *service.SyncService
	logger  *zap.Logger
}

func NewSyncHandler(s *service.SyncService, logger *zap.Logger) *SyncHandler {
	return &SyncHandler{service: s, logger: logger}
}

// GeneratePreview - preview symbols for a canonical without DB insert
func (h *SyncHandler) GeneratePreview(c *gin.Context) {
	var req domain.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Generate(c.Request.Context(), req.Canonical)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// BulkGeneratePreview - preview symbols from list of binance symbols
func (h *SyncHandler) BulkGeneratePreview(c *gin.Context) {
	var req domain.BulkGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.BulkGenerate(c.Request.Context(), req.Symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *SyncHandler) SyncBinance(c *gin.Context) {
	result, err := h.service.SyncBinance(c.Request.Context())
	if err != nil {
		h.logger.Error("Binance sync failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SyncHandler) SyncAll(c *gin.Context) {
	result, err := h.service.SyncAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Sync all failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
