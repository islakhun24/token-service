# Futures Symbol Module

Production-ready Go module that fetches, normalizes, matches, and ranks futures trading pairs across 6 major exchanges.

## Features

- **Multi-Exchange Support**: Binance, OKX, Bybit, Bitget, MEXC, KuCoin
- **Canonical Normalization**: All symbols parsed to `Base-Quote-PER` format
- **Automatic Cross-Exchange Matching**: No hardcoded mappings
- **Duplicate-Safe CoinGecko Resolution**: Scoring-based ID resolver
- **Market Cap Enrichment**: Cached CoinGecko market data
- **Concurrent & Thread-Safe**: Goroutines + mutex-protected caches
- **Clean Architecture**: Modular, scalable, production-ready

## API

```
GET /futures/pairs
```

Returns the strict JSON output format with ranked, enriched futures pairs.

## Project Structure

```
/internal
  /symbol       - Core domain: model, parser, validator, matcher, registry, ranker, service
  /collector    - Exchange-specific collectors (6 exchanges)
  /asset        - CoinGecko integration with duplicate-safe resolver
  /worker       - Background cache refresh workers
  /handler      - HTTP handlers
  /config       - Application configuration
/pkg
  /httpclient   - Production HTTP client
/cmd/server    - Application entry point
```

## Quick Start

```bash
# 1. Navigate to project
cd futures-symbol-module

# 2. Download dependencies
go mod tidy

# 3. Run server
go run ./cmd/server

# 4. Test endpoint
curl http://localhost:8080/futures/pairs
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT     | 8080    | HTTP server port |

## Architecture Highlights

### Symbol Parsing
Supports all major exchange formats:
- `BTCUSDT` (Binance/Bybit)
- `BTC-USDT-SWAP` (OKX)
- `BTCUSDT_UMCBL` (Bitget)
- `BTC_USDT` (MEXC)
- `BTC-USDTM` (KuCoin)

### Duplicate-Safe CoinGecko Resolver
When CoinGecko has multiple coins with the same symbol, the resolver scores each candidate:
- +0.5 symbol match
- +0.2 name relevance
- +0.3 known major coins (bitcoin, ethereum, etc.)
- -0.5 penalty for wrapped/bridged tokens
- Tie-break: shortest ID → first occurrence

### Caching Strategy
- `/coins/list` → refreshed every 6 hours
- `/coins/markets` → refreshed every 60 seconds
- No API calls per HTTP request
