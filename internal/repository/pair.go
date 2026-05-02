package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Pair represents the database schema for stored cryptocurrency pairs.
type Pair struct {
	Base        string  `gorm:"primaryKey;column:base" json:"base"`
	Quote       string  `gorm:"primaryKey;column:quote" json:"quote"`
	Type        string  `gorm:"primaryKey;column:type" json:"type"`
	Binance     *string `json:"binance"`
	Bybit       *string `json:"bybit"`
	OKX         *string `json:"okx"`
	Mexc        *string `json:"mexc"`
	Bitget      *string `json:"bitget"`
	Kucoin      *string `json:"kucoin"`
	CoinGeckoID *string `json:"coingecko_id"`
	MarketCap   float64 `json:"market_cap"`
}

// PairRepository defines operations for the Pair model.
type PairRepository struct {
	db *gorm.DB
}

// NewPairRepository creates a new PairRepository.
func NewPairRepository(db *gorm.DB) *PairRepository {
	return &PairRepository{db: db}
}

// BulkUpsert inserts or updates multiple pairs in the database.
func (r *PairRepository) BulkUpsert(ctx context.Context, pairs []Pair) error {
	if len(pairs) == 0 {
		return nil
	}

	// Use ON CONFLICT to update the existing rows based on the primary keys
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "base"}, {Name: "quote"}, {Name: "type"}},
		DoUpdates: clause.AssignmentColumns([]string{"binance", "bybit", "okx", "mexc", "bitget", "kucoin", "coin_gecko_id", "market_cap"}),
	}).CreateInBatches(pairs, 100).Error
}

// GetAll fetches all pairs from the database.
func (r *PairRepository) GetAll(ctx context.Context) ([]Pair, error) {
	var pairs []Pair
	if err := r.db.WithContext(ctx).Find(&pairs).Error; err != nil {
		return nil, err
	}
	return pairs, nil
}
