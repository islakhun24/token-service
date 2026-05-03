package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"futures-symbol-module/internal/repository"
)

// MCap tier thresholds
const (
	megaCapThreshold  = 200_000_000_000 // > $200B
	largeCapThreshold = 10_000_000_000  // $10B - $200B
	midCapThreshold   = 1_000_000_000   // $1B - $10B
	smallCapThreshold = 100_000_000     // $100M - $1B
	// microCap < $100M
)

// PairResponse extends Pair with computed fields for the API response.
type PairResponse struct {
	Base              string   `json:"base"`
	Quote             string   `json:"quote"`
	Type              string   `json:"type"`
	Binance           *string  `json:"binance"`
	Bybit             *string  `json:"bybit"`
	OKX               *string  `json:"okx"`
	Mexc              *string  `json:"mexc"`
	Bitget            *string  `json:"bitget"`
	Kucoin            *string  `json:"kucoin"`
	CMCID             *int     `json:"coinmarketcap_id"`
	Name              *string  `json:"name"`
	Slug              *string  `json:"slug"`
	CMCRank           *int     `json:"cmc_rank"`
	MarketCap         float64  `json:"market_cap"`
	CirculatingSupply float64  `json:"circulating_supply"`
	TotalSupply       float64  `json:"total_supply"`
	MaxSupply         float64  `json:"max_supply"`
	Categories        []string `json:"categories"`
	McapTier          string   `json:"mcap_tier"`
}

// DBPairsHandler handles endpoints for pairs fetched from the database.
type DBPairsHandler struct {
	repo *repository.PairRepository
}

// NewDBPairsHandler creates a new db pairs handler.
func NewDBPairsHandler(repo *repository.PairRepository) *DBPairsHandler {
	return &DBPairsHandler{repo: repo}
}

// GetPairs handles GET /api/db/pairs with optional filters and pagination.
// Query params: category, mcap_tier, page (default 1), limit (default 20, max 100)
func (h *DBPairsHandler) GetPairs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	params := repository.PairFilterParams{
		Category: c.Query("category"),
		McapTier: c.Query("mcap_tier"),
		Page:     page,
		Limit:    limit,
	}

	result, err := h.repo.GetPairsFiltered(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch pairs from database"})
		return
	}

	// Get categories map for enrichment
	categoriesMap, err := h.repo.GetCategoriesMap(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}

	// Build response with categories and mcap_tier
	var response []PairResponse
	for _, p := range result.Pairs {
		cats := categoriesMap[p.Base]
		if cats == nil {
			cats = []string{}
		}

		resp := PairResponse{
			Base:              p.Base,
			Quote:             p.Quote,
			Type:              p.Type,
			Binance:           p.Binance,
			Bybit:             p.Bybit,
			OKX:               p.OKX,
			Mexc:              p.Mexc,
			Bitget:            p.Bitget,
			Kucoin:            p.Kucoin,
			CMCID:             p.CMCID,
			Name:              p.CMCName,
			Slug:              p.CMCSlug,
			CMCRank:           p.CMCRank,
			MarketCap:         p.MarketCap,
			CirculatingSupply: p.CirculatingSupply,
			TotalSupply:       p.TotalSupply,
			MaxSupply:         p.MaxSupply,
			Categories:        cats,
			McapTier:          classifyMarketCap(p.MarketCap),
		}
		response = append(response, resp)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        response,
		"total":       result.Total,
		"page":        result.Page,
		"limit":       result.Limit,
		"total_pages": result.TotalPages,
	})
}

// GetCategories handles GET /api/categories.
func (h *DBPairsHandler) GetCategories(c *gin.Context) {
	categories, err := h.repo.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(categories),
		"data":  categories,
	})
}

// classifyMarketCap returns the market cap tier for a given market cap value.
func classifyMarketCap(marketCap float64) string {
	switch {
	case marketCap > megaCapThreshold:
		return "megacap"
	case marketCap >= largeCapThreshold:
		return "largecap"
	case marketCap >= midCapThreshold:
		return "midcap"
	case marketCap >= smallCapThreshold:
		return "smallcap"
	case marketCap > 0:
		return "microcap"
	default:
		return "unknown"
	}
}

