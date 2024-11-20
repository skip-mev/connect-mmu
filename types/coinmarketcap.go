package types

type CoinMarketCapInfo struct {
	// BaseID is the ID of a base asset on coinmarketcap
	BaseID int64 `json:"base_id"`
	// QuoteID is the ID of a quote asset on coinmarketcap
	QuoteID int64 `json:"quote_id"`

	BaseRank  int64 `json:"base_rank"`
	QuoteRank int64 `json:"quote_rank"`
}

func NewCoinMarketCapInfo(baseID, quoteID, baseRank, quoteRank int64) CoinMarketCapInfo {
	return CoinMarketCapInfo{
		BaseID:    baseID,
		QuoteID:   quoteID,
		BaseRank:  baseRank,
		QuoteRank: quoteRank,
	}
}

func (c *CoinMarketCapInfo) Invert() {
	base := c.BaseID
	quote := c.QuoteID
	c.BaseID = quote
	c.QuoteID = base

	baseRank := c.BaseRank
	quoteRank := c.QuoteRank
	c.QuoteRank = baseRank
	c.BaseRank = quoteRank
}

func (c *CoinMarketCapInfo) HasRank() bool {
	return c.BaseRank > 0
}
