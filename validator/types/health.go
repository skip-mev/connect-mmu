package types

// Counts keeps track of successes and failures seen in the logs for a currency_pair/provider pair.
type Counts struct {
	Success      int     `json:"success"`
	Failure      int     `json:"failure"`
	AveragePrice float64 `json:"average_price"`
}

// MarketHealth = ticker -> provider_name -> Count
type MarketHealth map[string]ProviderCounts

type ProviderCounts map[string]*Counts
