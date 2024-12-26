package provider

import (
	"fmt"
	"strings"
)

type Document struct {
	AssetInfos      []AssetInfo      `json:"asset_infos"`
	ProviderMarkets []ProviderMarket `json:"provider_markets"`
}

type AssetInfo struct {
	ID             int32      `json:"id"`
	Symbol         string     `json:"symbol"`
	IsCrypto       bool       `json:"is_crypto"`
	Rank           int64      `json:"rank"`
	RankValid      bool       `json:"rank_valid"`
	CMCID          int64      `json:"cmc_id"`
	CMCIDValid     bool       `json:"cmc_id_valid"`
	MultiAddresses [][]string `json:"multi_addresses"`
}

//nolint:revive
type ProviderMarket struct {
	ID               int32   `json:"id"`
	TargetBase       string  `json:"target_base"`
	TargetQuote      string  `json:"target_quote"`
	OffChainTicker   string  `json:"off_chain_ticker"`
	ProviderName     string  `json:"provider_name"`
	QuoteVolume      float64 `json:"quote_volume"`
	UsdVolume        float64 `json:"usd_volume"`
	BaseAssetInfoID  int32   `json:"base_asset_info_id"`
	QuoteAssetInfoID int32   `json:"quote_asset_info_id"`
	MetadataJSON     string  `json:"metadata_json"`
	ReferencePrice   float64 `json:"reference_price"`
	NegativeDepthTwo float64 `json:"negative_depth_two"`
	PositiveDepthTwo float64 `json:"positive_depth_two"`
}

// CreateProviderMarket wraps generated CreateProviderMarketParams with extra info.
type CreateProviderMarket struct {
	Create       CreateProviderMarketParams
	BaseAddress  string
	QuoteAddress string
}

func (cpm *CreateProviderMarket) ValidateBasic() error {
	if cpm.BaseAddress == "" && cpm.QuoteAddress != "" {
		return fmt.Errorf("baseAddress must be non-empty if quoteAddress is not")
	}

	if cpm.BaseAddress != "" && cpm.QuoteAddress == "" {
		return fmt.Errorf("baseAddress must be empty if quoteAddress is")
	}

	return cpm.Create.ValidateBasic()
}

// ValidateBasic performs basic validation on a CreateProviderMarketParams.
func (pm *CreateProviderMarketParams) ValidateBasic() error {
	if pm.TargetBase == "" {
		return fmt.Errorf("target base cannot be empty")
	}

	if pm.TargetQuote == "" {
		return fmt.Errorf("target quote cannot be empty")
	}

	if strings.ToUpper(pm.TargetBase) != pm.TargetBase {
		return fmt.Errorf("incorrectly formatted base string, expected: %s got: %s", strings.ToUpper(pm.TargetBase), pm.TargetBase)
	}
	if strings.ToUpper(pm.TargetQuote) != pm.TargetQuote {
		return fmt.Errorf("incorrectly formatted quote string, expected: %s got: %s", strings.ToUpper(pm.TargetQuote), pm.TargetQuote)
	}

	if pm.OffChainTicker == "" {
		return fmt.Errorf("offchain ticker cannot be empty")
	}

	if pm.ProviderName == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if pm.QuoteVolume < 0 {
		return fmt.Errorf("quote volume cannot be less than 0")
	}

	return nil
}

type CreateAssetInfoParams struct {
	Symbol         string
	CmcID          int64
	Rank           int64
	MultiAddresses [][]string
}

type CreateProviderMarketParams struct {
	TargetBase       string
	TargetQuote      string
	OffChainTicker   string
	ProviderName     string
	QuoteVolume      float64
	UsdVolume        float64
	BaseAssetInfoID  int32
	QuoteAssetInfoID int32
	MetadataJSON     []byte
	ReferencePrice   float64
	NegativeDepthTwo float64
	PositiveDepthTwo float64
}

type GetFilteredProviderMarketsParams struct {
	ProviderNames []string
}

type GetFilteredProviderMarketsRow struct {
	TargetBase       string
	TargetQuote      string
	OffChainTicker   string
	ProviderName     string
	QuoteVolume      float64
	UsdVolume        float64
	MetadataJSON     []byte
	ReferencePrice   float64
	NegativeDepthTwo float64
	PositiveDepthTwo float64
	BaseCmcID        int64
	QuoteCmcID       int64
	BaseRank         int64
	QuoteRank        int64
}
