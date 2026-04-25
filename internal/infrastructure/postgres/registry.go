package postgres

import (
	"context"
	"database/sql"
	_ "strings"
	_ "time"

	"token-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

type RegistryRepository struct {
	db *sqlx.DB
}

func NewRegistryRepository(db *sqlx.DB) *RegistryRepository {
	return &RegistryRepository{db: db}
}

func (r *RegistryRepository) GetAll(ctx context.Context) ([]domain.CanonicalToken, error) {
	var tokens []domain.CanonicalToken
	err := r.db.SelectContext(ctx, &tokens, `
		SELECT id, canonical_symbol, coingecko_id, name, category, is_active, arkham_ready, created_at, updated_at
		FROM canonical_tokens
		ORDER BY canonical_symbol
	`)
	return tokens, err
}

func (r *RegistryRepository) GetByCanonical(ctx context.Context, symbol string) (*domain.CanonicalToken, error) {
	var token domain.CanonicalToken
	err := r.db.GetContext(ctx, &token, `
		SELECT id, canonical_symbol, coingecko_id, name, category, is_active, arkham_ready, created_at, updated_at
		FROM canonical_tokens WHERE canonical_symbol = $1
	`, symbol)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *RegistryRepository) Create(ctx context.Context, token *domain.CanonicalToken) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO canonical_tokens (canonical_symbol, coingecko_id, name, category, is_active, arkham_ready)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, token.CanonicalSymbol, token.CoinGeckoID, token.Name, token.Category, token.IsActive, token.ArkhamReady).Scan(
		&token.ID, &token.CreatedAt, &token.UpdatedAt,
	)
}

func (r *RegistryRepository) Update(ctx context.Context, token *domain.CanonicalToken) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE canonical_tokens
		SET coingecko_id = $1, name = $2, category = $3, is_active = $4, arkham_ready = $5, updated_at = NOW()
		WHERE canonical_symbol = $6
	`, token.CoinGeckoID, token.Name, token.Category, token.IsActive, token.ArkhamReady, token.CanonicalSymbol)
	return err
}

func (r *RegistryRepository) Delete(ctx context.Context, canonical string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM canonical_tokens WHERE canonical_symbol = $1`, canonical)
	return err
}

func (r *RegistryRepository) GetExchangeMaps(ctx context.Context, canonicalID int) ([]domain.ExchangeMapping, error) {
	var maps []domain.ExchangeMapping
	err := r.db.SelectContext(ctx, &maps, `
		SELECT id, canonical_id, exchange, market_type, exchange_symbol, base_asset, quote_asset, normalized_base, is_active
		FROM exchange_mappings WHERE canonical_id = $1
	`, canonicalID)
	return maps, err
}

func (r *RegistryRepository) SaveExchangeMap(ctx context.Context, m *domain.ExchangeMapping) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO exchange_mappings (canonical_id, exchange, market_type, exchange_symbol, base_asset, quote_asset, normalized_base, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (exchange, exchange_symbol, market_type) DO UPDATE SET
			canonical_id = EXCLUDED.canonical_id,
			base_asset = EXCLUDED.base_asset,
			quote_asset = EXCLUDED.quote_asset,
			normalized_base = EXCLUDED.normalized_base,
			is_active = EXCLUDED.is_active
	`, m.CanonicalID, m.Exchange, m.MarketType, m.ExchangeSymbol, m.BaseAsset, m.QuoteAsset, m.NormalizedBase, m.IsActive)
	return err
}

func (r *RegistryRepository) DeleteExchangeMaps(ctx context.Context, canonicalID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM exchange_mappings WHERE canonical_id = $1`, canonicalID)
	return err
}

func (r *RegistryRepository) GetContracts(ctx context.Context, canonicalID int) (map[string]string, error) {
	var rows []struct {
		Chain   string `db:"chain"`
		Address string `db:"contract_address"`
	}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT chain, contract_address FROM token_contracts WHERE canonical_id = $1
	`, canonicalID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, r := range rows {
		result[r.Chain] = r.Address
	}
	return result, nil
}

func (r *RegistryRepository) SaveContract(ctx context.Context, canonicalID int, chain, address string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO token_contracts (canonical_id, chain, contract_address)
		VALUES ($1, $2, $3)
		ON CONFLICT (canonical_id, chain) DO UPDATE SET contract_address = EXCLUDED.contract_address
	`, canonicalID, chain, address)
	return err
}

func (r *RegistryRepository) DeleteContracts(ctx context.Context, canonicalID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM token_contracts WHERE canonical_id = $1`, canonicalID)
	return err
}

func (r *RegistryRepository) GetStats(ctx context.Context) (*domain.TokenStats, error) {
	var stats domain.TokenStats

	err := r.db.GetContext(ctx, &stats.Total, `SELECT COUNT(*) FROM canonical_tokens`)
	if err != nil {
		return nil, err
	}

	var active, arkhamReady int
	r.db.GetContext(ctx, &active, `SELECT COUNT(*) FROM canonical_tokens WHERE is_active = true`)
	r.db.GetContext(ctx, &arkhamReady, `SELECT COUNT(*) FROM canonical_tokens WHERE arkham_ready = true`)
	stats.Active = active
	stats.ArkhamReady = arkhamReady

	var multi int
	r.db.GetContext(ctx, &multi, `
		SELECT COUNT(*) FROM (
			SELECT canonical_id FROM exchange_mappings
			GROUP BY canonical_id HAVING COUNT(*) > 2
		) t
	`)
	stats.MultiExchange = multi

	var alias int
	r.db.GetContext(ctx, &alias, `
		SELECT COUNT(*) FROM exchange_mappings
		WHERE base_asset != normalized_base
	`)
	stats.Aliases = alias

	return &stats, nil
}

func (r *RegistryRepository) ToggleActive(ctx context.Context, canonical string, active bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE canonical_tokens SET is_active = $1, updated_at = NOW() WHERE canonical_symbol = $2
	`, active, canonical)
	return err
}
