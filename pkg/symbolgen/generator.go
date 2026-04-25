package symbolgen

import (
	_ "fmt"
	"strings"
)

// Generator creates exchange-specific symbols from canonical name
type Generator struct {
	aliases map[string]string
}

func NewGenerator() *Generator {
	return &Generator{
		aliases: map[string]string{
			// KuCoin uses XBT for BTC
			"kucoin:BTC":  "XBT",
			"kucoin:ETH":  "ETH",
			"kucoin:SOL":  "SOL",
			"kucoin:XRP":  "XRP",
			"kucoin:DOGE": "DOGE",

			// Binance uses 1000 prefix for small coins
			"binance:PEPE":  "1000PEPE",
			"binance:SHIB":  "1000SHIB",
			"binance:FLOKI": "1000FLOKI",
			"binance:BONK":  "1000BONK",
			"binance:LUNC":  "1000LUNC",
		},
	}
}

func (g *Generator) AddAlias(exchange, canonical, exchangeBase string) {
	key := strings.ToLower(exchange) + ":" + strings.ToUpper(canonical)
	g.aliases[key] = strings.ToUpper(exchangeBase)
}

func (g *Generator) ForBinance(canonical string) string {
	canonical = strings.ToUpper(canonical)
	if alt, ok := g.aliases["binance:"+canonical]; ok {
		return alt + "USDT"
	}
	return canonical + "USDT"
}

func (g *Generator) ForKuCoin(canonical string) string {
	canonical = strings.ToUpper(canonical)
	base := canonical
	if alt, ok := g.aliases["kucoin:"+canonical]; ok {
		base = alt
	}
	return base + "USDTM"
}

func (g *Generator) ForBybit(canonical string) string {
	return strings.ToUpper(canonical) + "USDT"
}

func (g *Generator) ForOKX(canonical string) string {
	return strings.ToUpper(canonical) + "-USDT-SWAP"
}

func (g *Generator) ForArkham(canonical string) string {
	return strings.ToLower(guessCoinGeckoID(canonical))
}

func (g *Generator) ForCoinGecko(canonical string) string {
	return g.ForArkham(canonical)
}

func (g *Generator) GenerateAll(canonical string) *PlatformSymbols {
	return &PlatformSymbols{
		Canonical: strings.ToUpper(canonical),
		Binance:   g.ForBinance(canonical),
		KuCoin:    g.ForKuCoin(canonical),
		Bybit:     g.ForBybit(canonical),
		OKX:       g.ForOKX(canonical),
		Arkham:    g.ForArkham(canonical),
		CoinGecko: g.ForCoinGecko(canonical),
	}
}

type PlatformSymbols struct {
	Canonical string `json:"canonical"`
	Binance   string `json:"binance"`
	KuCoin    string `json:"kucoin"`
	Bybit     string `json:"bybit"`
	OKX       string `json:"okx"`
	Arkham    string `json:"arkham"`
	CoinGecko string `json:"coingecko"`
}

func guessCoinGeckoID(canonical string) string {
	canonical = strings.ToUpper(canonical)
	known := map[string]string{
		"BTC": "bitcoin", "ETH": "ethereum", "BNB": "binancecoin",
		"SOL": "solana", "XRP": "ripple", "ADA": "cardano",
		"DOGE": "dogecoin", "DOT": "polkadot", "AVAX": "avalanche-2",
		"MATIC": "matic-network", "LINK": "chainlink", "UNI": "uniswap",
		"ATOM": "cosmos", "ETC": "ethereum-classic", "LTC": "litecoin",
		"BCH": "bitcoin-cash", "ALGO": "algorand", "VET": "vechain",
		"FIL": "filecoin", "TRX": "tron", "NEAR": "near",
		"APT": "aptos", "ARB": "arbitrum", "OP": "optimism",
		"SUI": "sui", "SEI": "sei-network", "INJ": "injective-protocol",
		"RUNE": "thorchain", "PEPE": "pepe", "SHIB": "shiba-inu",
		"BONK": "bonk", "FLOKI": "floki", "DOGS": "dogs-2",
		"WIF": "dogwifhat", "BOME": "book-of-meme", "WLD": "worldcoin-wld",
		"ARKM": "arkham", "PYTH": "pyth-network", "JTO": "jito",
		"JUP": "jupiter-exchange-solana", "TIA": "celestia",
		"DYM": "dymension", "STRK": "starknet", "ZRO": "layerzero",
		"ENA": "ethena", "W": "wormhole", "TNSR": "tensor",
		"PRCL": "parcl", "DRIFT": "drift-protocol", "ZEX": "zeta",
		"CLORE": "clore-ai", "RATS": "rats", "SATS": "sats-ordinals",
		"ORDI": "ordinals", "PIZZA": "pizza-2", "LUNC": "terra-luna",
		"LUNA": "terra-luna-2", "FTT": "ftx-token", "COTI": "coti",
		"BEAM": "beam-2", "KAS": "kaspa", "RENDER": "render-token",
		"FET": "fetch-ai", "AGIX": "singularitynet", "OCEAN": "ocean-protocol",
		"NMR": "numeraire", "LPT": "livepeer", "GMT": "stepn",
		"GALA": "gala", "IMX": "immutable-x", "DYDX": "dydx",
		"CRV": "curve-dao-token", "APE": "apecoin", "SAND": "the-sandbox",
		"MANA": "decentraland", "AXS": "axie-infinity", "CHZ": "chiliz",
		"ENJ": "enjincoin", "BAT": "basic-attention-token",
		"COMP": "compound-governance-token", "MKR": "maker",
		"AAVE": "aave", "1INCH": "1inch", "SUSHI": "sushi",
		"YFI": "yearn-finance", "SNX": "havven", "BAL": "balancer",
		"KNC": "kyber-network-crystal", "GRT": "the-graph",
		"LRC": "loopring", "ZRX": "0x", "REN": "republic-protocol",
		"UMA": "uma", "BAND": "band-protocol", "RLC": "iexec-rlc",
		"STORJ": "storj", "ANT": "aragon", "OXT": "orchid-protocol",
		"KEEP": "keep-network", "NU": "nucypher", "BNT": "bancor",
		"MLN": "melon", "REP": "augur", "LEND": "ethlend",
		"AMPL": "ampleforth", "PAXG": "pax-gold", "XAUT": "tether-gold",
	}
	if id, ok := known[canonical]; ok {
		return id
	}
	return strings.ToLower(canonical)
}
