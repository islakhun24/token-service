package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Pair represents the database schema for stored cryptocurrency pairs.
type Pair struct {
	Base               string  `gorm:"primaryKey;column:base" json:"base"`
	Quote              string  `gorm:"primaryKey;column:quote" json:"quote"`
	Type               string  `gorm:"primaryKey;column:type" json:"type"`
	Binance            *string `json:"binance"`
	Bybit              *string `json:"bybit"`
	OKX                *string `json:"okx"`
	Mexc               *string `json:"mexc"`
	Bitget             *string `json:"bitget"`
	Kucoin             *string `json:"kucoin"`
	CMCID              *int    `gorm:"column:cmc_id" json:"coinmarketcap_id"`
	CMCName            *string `gorm:"column:cmc_name" json:"name"`
	CMCSlug            *string `gorm:"column:cmc_slug" json:"slug"`
	CMCRank            *int    `gorm:"column:cmc_rank" json:"cmc_rank"`
	MarketCap          float64 `json:"market_cap"`
	CirculatingSupply  float64 `gorm:"column:circulating_supply" json:"circulating_supply"`
	TotalSupply        float64 `gorm:"column:total_supply" json:"total_supply"`
	MaxSupply          float64 `gorm:"column:max_supply" json:"max_supply"`
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
		DoUpdates: clause.AssignmentColumns([]string{"binance", "bybit", "okx", "mexc", "bitget", "kucoin", "cmc_id", "cmc_name", "cmc_slug", "cmc_rank", "market_cap", "circulating_supply", "total_supply", "max_supply"}),
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
