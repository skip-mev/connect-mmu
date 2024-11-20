package types

import (
	"math/big"
	"strconv"

	"github.com/skip-mev/connect/v2/x/marketmap/types/tickermetadata"

	"github.com/skip-mev/connect-mmu/types"
)

const (
	VenueCoinMarketcap = "coinmarketcap"
)

// ToTickerMetadataJSON creates a JSON string from the given database row based on the chain
// type of this generation run.
func ToTickerMetadataJSON(feed Feed, referencePrice *big.Float, totalLiquidity float64) (string, error) {
	// scale the price by decimals
	md := tickermetadata.DyDx{
		ReferencePrice: types.ScalePriceToUint64(referencePrice),
		Liquidity:      uint64(totalLiquidity),
		AggregateIDs:   make([]tickermetadata.AggregatorID, 0),
	}

	// Base Asset
	md.AggregateIDs = append(md.AggregateIDs, tickermetadata.AggregatorID{
		Venue: VenueCoinMarketcap,
		ID:    strconv.FormatInt(feed.CMCInfo.BaseID, 10),
	})

	bz, err := tickermetadata.MarshalDyDx(md)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
