package marketmap

import (
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	slinkytypes "github.com/skip-mev/slinky/pkg/types"
	slinkymmtypes "github.com/skip-mev/slinky/x/marketmap/types"
)

func SlinkyToConnectMarket(market slinkymmtypes.Market) mmtypes.Market {
	convertedProviderConfigs := make([]mmtypes.ProviderConfig, 0)

	for _, providerConfig := range market.ProviderConfigs {
		convertedProviderConfig := mmtypes.ProviderConfig{
			Name:           providerConfig.Name,
			OffChainTicker: providerConfig.OffChainTicker,
			Invert:         providerConfig.Invert,
			Metadata_JSON:  providerConfig.Metadata_JSON,
		}

		if providerConfig.NormalizeByPair != nil {
			convertedProviderConfig.NormalizeByPair = &connecttypes.CurrencyPair{
				Base:  providerConfig.NormalizeByPair.Base,
				Quote: providerConfig.NormalizeByPair.Quote,
			}
		}

		convertedProviderConfigs = append(convertedProviderConfigs, convertedProviderConfig)
	}

	newMarket := mmtypes.Market{
		Ticker: mmtypes.Ticker{
			CurrencyPair: connecttypes.CurrencyPair{
				Base:  market.Ticker.CurrencyPair.Base,
				Quote: market.Ticker.CurrencyPair.Quote,
			},
			Decimals:         market.Ticker.Decimals,
			MinProviderCount: market.Ticker.MinProviderCount,
			Enabled:          market.Ticker.Enabled,
			Metadata_JSON:    market.Ticker.Metadata_JSON,
		},
		ProviderConfigs: convertedProviderConfigs,
	}

	return newMarket
}

func SlinkyToConnectMarkets(markets []slinkymmtypes.Market) []mmtypes.Market {
	convertedMarkets := make([]mmtypes.Market, len(markets))
	for i, market := range markets {
		convertedMarkets[i] = SlinkyToConnectMarket(market)
	}

	return convertedMarkets
}

func SlinkyToConnectMarketMap(marketMap slinkymmtypes.MarketMap) mmtypes.MarketMap {
	mm := mmtypes.MarketMap{
		Markets: make(map[string]mmtypes.Market),
	}

	for _, market := range marketMap.Markets {
		newMarket := SlinkyToConnectMarket(market)
		mm.Markets[newMarket.Ticker.String()] = newMarket
	}

	return mm
}

func ConnectToSlinkyMarket(market mmtypes.Market) slinkymmtypes.Market {
	convertedProviderConfigs := make([]slinkymmtypes.ProviderConfig, 0)

	for _, providerConfig := range market.ProviderConfigs {
		convertedProviderConfig := slinkymmtypes.ProviderConfig{
			Name:           providerConfig.Name,
			OffChainTicker: providerConfig.OffChainTicker,
			Invert:         providerConfig.Invert,
			Metadata_JSON:  providerConfig.Metadata_JSON,
		}

		if providerConfig.NormalizeByPair != nil {
			convertedProviderConfig.NormalizeByPair = &slinkytypes.CurrencyPair{
				Base:  providerConfig.NormalizeByPair.Base,
				Quote: providerConfig.NormalizeByPair.Quote,
			}
		}

		convertedProviderConfigs = append(convertedProviderConfigs, convertedProviderConfig)
	}

	newMarket := slinkymmtypes.Market{
		Ticker: slinkymmtypes.Ticker{
			CurrencyPair: slinkytypes.CurrencyPair{
				Base:  market.Ticker.CurrencyPair.Base,
				Quote: market.Ticker.CurrencyPair.Quote,
			},
			Decimals:         market.Ticker.Decimals,
			MinProviderCount: market.Ticker.MinProviderCount,
			Enabled:          market.Ticker.Enabled,
			Metadata_JSON:    market.Ticker.Metadata_JSON,
		},
		ProviderConfigs: convertedProviderConfigs,
	}

	return newMarket
}

func ConnectToSlinkyMarkets(markets []mmtypes.Market) []slinkymmtypes.Market {
	convertedMarkets := make([]slinkymmtypes.Market, len(markets))
	for i, market := range markets {
		convertedMarkets[i] = ConnectToSlinkyMarket(market)
	}

	return convertedMarkets
}

func ConnectToSlinkyMarketMap(marketMap mmtypes.MarketMap) slinkymmtypes.MarketMap {
	mm := slinkymmtypes.MarketMap{
		Markets: make(map[string]slinkymmtypes.Market),
	}

	for _, market := range marketMap.Markets {
		newMarket := ConnectToSlinkyMarket(market)

		mm.Markets[newMarket.Ticker.String()] = newMarket
	}

	return mm
}
