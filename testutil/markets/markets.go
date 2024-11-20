package markets

import (
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
)

var UsdtUsd = mmtypes.Market{
	Ticker: mmtypes.Ticker{
		CurrencyPair: connecttypes.CurrencyPair{
			Base:  "USDT",
			Quote: "USD",
		},
		Decimals:         8,
		MinProviderCount: 1,
		Enabled:          true,
	},
	ProviderConfigs: []mmtypes.ProviderConfig{
		{
			Name:           "okx_ws",
			OffChainTicker: "USDC-USDT",
			Invert:         true,
		},
	},
}
