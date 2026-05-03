package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PairCategory represents the many-to-many relationship between pairs and categories.
type PairCategory struct {
	Base     string `gorm:"primaryKey;column:base" json:"base"`
	Category string `gorm:"primaryKey;column:category" json:"category"`
}

// TableName overrides the default table name.
func (PairCategory) TableName() string {
	return "pair_categories"
}

// BulkUpsertCategories inserts categories for a given base coin, ignoring duplicates.
func (r *PairRepository) BulkUpsertCategories(ctx context.Context, base string, categories []string) error {
	if len(categories) == 0 {
		return nil
	}

	var records []PairCategory
	for _, cat := range categories {
		records = append(records, PairCategory{
			Base:     base,
			Category: cat,
		})
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(records, 100).Error
}

// GetAllCategories returns all distinct category names from the database.
func (r *PairRepository) GetAllCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).
		Model(&PairCategory{}).
		Distinct("category").
		Order("category ASC").
		Pluck("category", &categories).Error
	return categories, err
}

// GetCategoriesForBase returns categories for a given base coin.
func (r *PairRepository) GetCategoriesForBase(ctx context.Context, base string) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).
		Model(&PairCategory{}).
		Where("base = ?", base).
		Pluck("category", &categories).Error
	return categories, err
}

// GetCategoriesMap returns a map of base -> []category for all coins.
func (r *PairRepository) GetCategoriesMap(ctx context.Context) (map[string][]string, error) {
	var records []PairCategory
	if err := r.db.WithContext(ctx).Find(&records).Error; err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, r := range records {
		result[r.Base] = append(result[r.Base], r.Category)
	}
	return result, nil
}

// GetBasesWithCategories returns all base symbols that already have categories assigned.
func (r *PairRepository) GetBasesWithCategories(ctx context.Context) (map[string]bool, error) {
	var bases []string
	err := r.db.WithContext(ctx).
		Model(&PairCategory{}).
		Distinct("base").
		Pluck("base", &bases).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(bases))
	for _, b := range bases {
		result[b] = true
	}
	return result, nil
}



// PairFilterParams holds filter and pagination parameters.
type PairFilterParams struct {
	Category string
	McapTier string
	Page     int
	Limit    int
}

// PairFilterResult holds the paginated result.
type PairFilterResult struct {
	Pairs      []Pair
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

// GetPairsFiltered fetches pairs with optional category, mcap_tier filters and pagination.
func (r *PairRepository) GetPairsFiltered(ctx context.Context, params PairFilterParams) (*PairFilterResult, error) {
	query := r.db.WithContext(ctx).Model(&Pair{})

	// Filter by category via subquery on pair_categories
	if params.Category != "" {
		query = query.Where("base IN (?)",
			r.db.Model(&PairCategory{}).
				Select("base").
				Where("LOWER(category) = LOWER(?)", params.Category),
		)
	}

	// Filter by market cap tier
	if params.McapTier != "" {
		switch params.McapTier {
		case "megacap":
			query = query.Where("market_cap > ?", 200_000_000_000)
		case "largecap":
			query = query.Where("market_cap >= ? AND market_cap <= ?", 10_000_000_000, 200_000_000_000)
		case "midcap":
			query = query.Where("market_cap >= ? AND market_cap < ?", 1_000_000_000, 10_000_000_000)
		case "smallcap":
			query = query.Where("market_cap >= ? AND market_cap < ?", 100_000_000, 1_000_000_000)
		case "microcap":
			query = query.Where("market_cap > 0 AND market_cap < ?", 100_000_000)
		}
	}

	// Count total matching records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	page := params.Page
	limit := params.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	var pairs []Pair
	if err := query.Order("market_cap DESC").Offset(offset).Limit(limit).Find(&pairs).Error; err != nil {
		return nil, err
	}

	return &PairFilterResult{
		Pairs:      pairs,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// DeleteCategoriesForBase removes all categories for a given base coin.
func (r *PairRepository) DeleteCategoriesForBase(ctx context.Context, base string) error {
	return r.db.WithContext(ctx).
		Where("base = ?", base).
		Delete(&PairCategory{}).Error
}

// ReplaceCategoriesForBase replaces all categories for a base coin (delete + insert).
func (r *PairRepository) ReplaceCategoriesForBase(ctx context.Context, base string, categories []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing
		if err := tx.Where("base = ?", base).Delete(&PairCategory{}).Error; err != nil {
			return err
		}

		if len(categories) == 0 {
			return nil
		}

		// Insert new
		var records []PairCategory
		for _, cat := range categories {
			records = append(records, PairCategory{
				Base:     base,
				Category: cat,
			})
		}

		return tx.CreateInBatches(records, 100).Error
	})
}
