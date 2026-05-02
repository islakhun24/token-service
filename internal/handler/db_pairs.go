package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"futures-symbol-module/internal/repository"
)

// DBPairsHandler handles endpoints for pairs fetched from the database.
type DBPairsHandler struct {
	repo *repository.PairRepository
}

// NewDBPairsHandler creates a new db pairs handler.
func NewDBPairsHandler(repo *repository.PairRepository) *DBPairsHandler {
	return &DBPairsHandler{repo: repo}
}

// GetPairs handles GET /api/db/pairs.
func (h *DBPairsHandler) GetPairs(c *gin.Context) {
	pairs, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pairs from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(pairs),
		"data":  pairs,
	})
}
