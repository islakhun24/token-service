package postgres

import (
	"fmt"

	"token-service/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDB(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
	return sqlx.Connect("postgres", dsn)
}

func Migrate(db *sqlx.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS canonical_tokens (
    id SERIAL PRIMARY KEY,
    canonical_symbol VARCHAR(20) UNIQUE NOT NULL,
    coingecko_id VARCHAR(100),
    name VARCHAR(200),
    category VARCHAR(50) DEFAULT 'unknown',
    is_active BOOLEAN DEFAULT true,
    arkham_ready BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchange_mappings (
    id SERIAL PRIMARY KEY,
    canonical_id INTEGER REFERENCES canonical_tokens(id) ON DELETE CASCADE,
    exchange VARCHAR(50) NOT NULL,
    market_type VARCHAR(50) NOT NULL,
    exchange_symbol VARCHAR(100) NOT NULL,
    base_asset VARCHAR(50) NOT NULL,
    quote_asset VARCHAR(50) NOT NULL,
    normalized_base VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    UNIQUE(exchange, exchange_symbol, market_type)
);

CREATE TABLE IF NOT EXISTS token_contracts (
    id SERIAL PRIMARY KEY,
    canonical_id INTEGER REFERENCES canonical_tokens(id) ON DELETE CASCADE,
    chain VARCHAR(50) NOT NULL,
    contract_address VARCHAR(100) NOT NULL,
    decimals INTEGER DEFAULT 18,
    UNIQUE(canonical_id, chain)
);

CREATE INDEX IF NOT EXISTS idx_exchange_maps_canonical ON exchange_mappings(canonical_id);
CREATE INDEX IF NOT EXISTS idx_exchange_maps_symbol ON exchange_mappings(exchange, exchange_symbol);
CREATE INDEX IF NOT EXISTS idx_contracts_canonical ON token_contracts(canonical_id);
`
	_, err := db.Exec(schema)
	return err
}
