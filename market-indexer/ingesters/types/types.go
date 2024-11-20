package types

const (
	// DefiTickerDelimiter is the delimiter used to delimit fields in a defi ticker.
	// Ex:
	//
	// $SWTS,solana,raydium,26mBPPJwru8qgqgSzRikqrdRsaFN8JdDP5jgszYmzyDX/SOL,solana,raydium,FamAVVw3t74CLufKYDfV9EBBTxefavHuLKC5Gdw7ZqJf.
	DefiTickerDelimiter = ","

	// TickerSeparator is the separator between BASE and QUOTE in a market pair ticker.
	TickerSeparator = "/"

	ProviderNameSuffixAPI = "_api"
	ProviderNameSuffixWS  = "_ws"
)
