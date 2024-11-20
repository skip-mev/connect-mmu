package huobi

var (
	knownQuotes = []string{
		"btc",
		"usd",
		"usdt",
		"eth",
		"bnb",
		"pax",
		"usdc",
		"xrp",
		"usds",
		"trx",
		"ngn",
		"rub",
		"try",
		"eur",
		"zar",
		"usdd",
	}

	ignoreSymbols = map[string]struct{}{
		"y1y2x8z9":   {},
		"h6h6x8z9":   {},
		"a8s83sd8f8": {},
		"q1w2x8z9":   {},
		"h4t4x8z9":   {},
		"h1t1x6z7":   {},
		"h3h3x8z9":   {},
		"c1d1x6z7":   {},
		"h1t1x8z9":   {},
		"h3t3x6z7":   {},
		"h2t2x6z7":   {},
		"h3t3x8z9":   {},
		"h2t2x8z9":   {},
		"h4t4x6z7":   {},
		"e9f9x8z9":   {},
		"a3b4x6z7":   {},
		"a3b4x8z9":   {},
		"h6h6x6z7":   {},
		"w4j4x8z9":   {},
		"h8h8x6z7":   {},
		"a8s83ld8f8": {},
		"h9h9x6z7":   {},
		"a8s8d8f8":   {},
		"f1f2x8z9":   {},
	}
)
