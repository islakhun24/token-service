# Token Service

Microservice untuk canonical token management, exchange symbol normalization, dan multi-platform token registry.

## Features

- **Canonical Token Registry**: Satu token = satu canonical symbol, map ke N exchange
- **Symbol Normalization**: XBT → BTC, 1000PEPE → PEPE, dll
- **Auto Symbol Generation**: Dari Binance symbol, auto-generate KuCoin/Bybit/OKX/Arkham/CoinGecko symbols
- **Multi-Exchange Support**: Binance, KuCoin, Bybit, OKX
- **Dynamic Sync**: Fetch Binance → generate all platforms → upsert ke DB
- **PostgreSQL Persistence**: Token data persisten dengan migrations

## Quick Start

```bash
# 1. Setup PostgreSQL
createdb token_service

# 2. Install dependencies
go mod tidy

# 3. Run service
go run cmd/server/main.go

# 4. Generate preview (tanpa insert DB)
curl -X POST http://localhost:8081/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{"canonical":"BTC"}'

# 5. Sync dari Binance (auto-generate semua platform + insert DB)
curl -X POST http://localhost:8081/api/v1/sync/binance
```

## API Endpoints

### Token Registry
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tokens` | List all tokens |
| POST | `/api/v1/tokens` | Create token manual |
| GET | `/api/v1/tokens/:canonical` | Get token detail |
| PUT | `/api/v1/tokens/:canonical` | Update token |
| DELETE | `/api/v1/tokens/:canonical` | Delete token |
| POST | `/api/v1/tokens/:canonical/verify` | Verify Arkham |
| POST | `/api/v1/tokens/:canonical/toggle` | Toggle active |
| GET | `/api/v1/exchanges/:exchange/symbols` | Get exchange symbols |
| GET | `/api/v1/normalize?exchange=&symbol=` | Normalize symbol |
| GET | `/api/v1/stats` | Registry stats |

### Symbol Generation (Preview Only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/generate` | Generate symbols untuk 1 token |
| POST | `/api/v1/generate/bulk` | Generate symbols untuk list Binance symbols |

### Dynamic Sync (Insert ke DB)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sync/binance` | Fetch Binance → generate all → upsert DB |
| POST | `/api/v1/sync/all` | Alias untuk sync/binance |

## Example: Generate Preview

```bash
# Generate symbols untuk BTC
curl -X POST http://localhost:8081/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{"canonical":"BTC"}'

# Response:
# {
#   "canonical": "BTC",
#   "binance": "BTCUSDT",
#   "kucoin": "XBTUSDTM",
#   "bybit": "BTCUSDT",
#   "okx": "BTC-USDT-SWAP",
#   "arkham": "bitcoin",
#   "coingecko": "bitcoin"
# }
```

## Example: Bulk Generate dari Binance Symbols

```bash
# Generate untuk list symbol
curl -X POST http://localhost:8081/api/v1/generate/bulk \
  -H "Content-Type: application/json" \
  -d '{"symbols":["BTCUSDT","ETHUSDT","PEPEUSDT","1000PEPEUSDT"]}'

# Response:
# {
#   "items": [
#     {"canonical":"BTC","binance":"BTCUSDT","kucoin":"XBTUSDTM",...},
#     {"canonical":"ETH","binance":"ETHUSDT","kucoin":"ETHUSDTM",...},
#     {"canonical":"PEPE","binance":"PEPEUSDT","kucoin":"PEPEUSDTM",...}
#   ],
#   "total": 3
# }
```

## Example: Sync Binance (Auto Insert DB)

```bash
# Fetch semua Binance perpetual → generate all platforms → upsert ke DB
curl -X POST http://localhost:8081/api/v1/sync/binance

# Response:
# {
#   "exchange": "binance",
#   "total": 312,
#   "inserted": 45,
#   "updated": 267,
#   "failed": 0,
#   "durationMs": 2340
# }
```

## How It Works

### 1. Fetch Binance Perpetuals
```
GET https://fapi.binance.com/fapi/v1/exchangeInfo
→ BTCUSDT, ETHUSDT, PEPEUSDT, 1000PEPEUSDT, XRPUSDT...
```

### 2. Normalize ke Canonical
```
BTCUSDT      → BTC
ETHUSDT      → ETH
PEPEUSDT     → PEPE
1000PEPEUSDT → PEPE  (alias: 1000PEPE → PEPE)
```

### 3. Generate All Platform Symbols
```
Canonical: BTC
  Binance:   BTCUSDT
  KuCoin:    XBTUSDTM     (alias: BTC → XBT)
  Bybit:     BTCUSDT
  OKX:       BTC-USDT-SWAP
  Arkham:    bitcoin
  CoinGecko: bitcoin

Canonical: PEPE
  Binance:   PEPEUSDT / 1000PEPEUSDT
  KuCoin:    PEPEUSDTM
  Bybit:     PEPEUSDT
  OKX:       PEPE-USDT-SWAP
  Arkham:    pepe
  CoinGecko: pepe
```

### 4. Upsert ke Database
```sql
-- Insert canonical token
INSERT INTO canonical_tokens (canonical_symbol, coingecko_id, name)
VALUES ('BTC', 'bitcoin', 'BTC')
ON CONFLICT (canonical_symbol) DO UPDATE SET coingecko_id = EXCLUDED.coingecko_id;

-- Insert exchange mappings untuk semua platform
INSERT INTO exchange_mappings (canonical_id, exchange, exchange_symbol, ...)
VALUES 
  (1, 'binance', 'BTCUSDT', ...),
  (1, 'kucoin', 'XBTUSDTM', ...),
  (1, 'bybit', 'BTCUSDT', ...),
  (1, 'okx', 'BTC-USDT-SWAP', ...)
ON CONFLICT (exchange, exchange_symbol, market_type) DO UPDATE SET ...;
```

## Project Structure

```
token-service/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── domain/
│   ├── handler/
│   │   ├── registry.go
│   │   └── sync.go
│   ├── service/
│   │   ├── registry.go
│   │   └── sync.go
│   └── infrastructure/
│       ├── binance/
│       ├── postgres/
│       └── staticmapping/
├── pkg/
│   ├── normalizer/         # Symbol normalization (XBT→BTC)
│   └── symbolgen/          # Auto-generate platform symbols
├── data/tokens.json
└── config.yaml
```
