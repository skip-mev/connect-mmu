package raydium

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	connectraydium "github.com/skip-mev/connect/v2/providers/apis/defi/raydium"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	raydium "github.com/skip-mev/connect-mmu/market-indexer/ingesters/raydium/generated/raydium_amm"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "raydium"
	ProviderName = Name + types.ProviderNameSuffixAPI

	// defaultRequestChunk is the size of the request that can be made to a solana node.
	defaultRequestChunk = 100
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the binance implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new raydium Ingester.
func New(logger *zap.Logger, cfg config.MarketConfig) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewClient(logger, cfg),
	}
}

func (ig *Ingester) GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error) {
	ig.logger.Info("fetching data")

	ig.logger.Info("querying token registry entries")

	tokenMetadata, err := ig.client.TokenMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch token metadata: %w", err)
	}

	symbolMap := make(map[string]string, len(tokenMetadata.Content))
	for _, entry := range tokenMetadata.Content {
		symbolMap[entry.Address] = entry.Symbol
	}

	ig.logger.Info("number of token entries", zap.Int("entries", len(symbolMap)))
	ig.logger.Info("querying pairs")

	pairs, err := ig.client.Pairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch pairs: %w", err)
	}

	ig.logger.Info("pairs", zap.Int("amount", len(pairs)))

	respAccounts, err := ig.chunkedRequests(ctx, pairs, defaultRequestChunk)
	if err != nil {
		ig.logger.Error("failed to run", zap.Error(err))
		return nil, err
	}

	pms := make([]provider.CreateProviderMarket, 0, len(respAccounts))

	ig.logger.Info("creating db entries")
	for i, acct := range respAccounts {
		pair := pairs[i]
		if acct == nil {
			continue
		}

		targetBase, targetQuote, err := getTargets(pair, symbolMap)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch target base and quote: %w", err)
		}

		var ammAccountData raydium.AmmInfo
		if err := bin.NewBinDecoder(acct.Data.GetBinary()).Decode(&ammAccountData); err != nil {
			continue
		}

		// construct connect metadata and perform basic validation
		meta := connectraydium.TickerMetadata{
			BaseTokenVault: connectraydium.AMMTokenVaultMetadata{
				TokenVaultAddress: ammAccountData.TokenCoin.String(),
				TokenDecimals:     ammAccountData.CoinDecimals,
			},
			QuoteTokenVault: connectraydium.AMMTokenVaultMetadata{
				TokenVaultAddress: ammAccountData.TokenPc.String(),
				TokenDecimals:     ammAccountData.PcDecimals,
			},
			AMMInfoAddress:    pair.AmmID,
			OpenOrdersAddress: ammAccountData.OpenOrders.String(),
		}

		if err := meta.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("invalid token metadata: %w", err)
		}

		bz, err := json.Marshal(meta)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal provider market metadata: %w", err)
		}

		ig.logger.Debug("pairs", zap.Any("meta", meta))

		targetBase, err = symbols.ToTickerString(targetBase)
		if err != nil {
			ig.logger.Debug("failed to convert target base to ticker string - skipping", zap.Error(err))
			continue
		}

		targetQuote, err = symbols.ToTickerString(targetQuote)
		if err != nil {
			ig.logger.Debug("failed to convert target quote to ticker string - skipping", zap.Error(err))
			continue
		}
		targetBaseOffchain := strings.ToUpper(strings.Join([]string{targetBase, Name, pair.BaseMint},
			types.DefiTickerDelimiter))
		targetQuoteOffchain := strings.ToUpper(strings.Join([]string{targetQuote, Name, pair.QuoteMint},
			types.DefiTickerDelimiter))

		cp := connecttypes.NewCurrencyPair(targetBaseOffchain, targetQuoteOffchain)
		if err := cp.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("invalid token metadata: %w", err)
		}

		pm := provider.CreateProviderMarket{
			Create: provider.CreateProviderMarketParams{
				TargetBase:       targetBase,
				TargetQuote:      targetQuote,
				OffChainTicker:   strings.Join([]string{targetBaseOffchain, targetQuoteOffchain}, types.TickerSeparator),
				ProviderName:     ProviderName,
				QuoteVolume:      pair.Volume24HQuote,
				UsdVolume:        pair.Volume24H,
				MetadataJSON:     bz,
				ReferencePrice:   pair.Price,
				PositiveDepthTwo: pair.Liquidity / 2,
				NegativeDepthTwo: pair.Liquidity / 2,
			},
			BaseAddress:  pair.BaseMint,
			QuoteAddress: pair.QuoteMint,
		}

		if err := pm.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("invalid provider market: %w: %v", err, pm)
		}

		pms = append(pms, pm)
	}

	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (ig *Ingester) Name() string {
	return Name
}

// getTargets gets the target base and quote from PairData and some known symbols.
// if the targets cannot be found, UNKNOWN is returned.
func getTargets(pair PairData, symbolMap map[string]string) (base, quote string, err error) {
	var ok bool

	// split pair names based on how they are formatted from the raydium API
	// pair.Name -> Base-QUOTE
	nameSplit := strings.Split(pair.Name, "-")

	// if the length of the split is not 2, we cannot reason about it
	// continue using UNKNOWN
	if len(nameSplit) != 2 {
		// fall back to using list of known quotes
		nameSplit = []string{symbols.TargetUnknown, symbols.TargetUnknown}
		for _, knownQuote := range knownQuotes {
			first, ok := strings.CutSuffix(pair.Name, "-"+"SOL")
			if ok {
				nameSplit = []string{first, knownQuote}
				break
			}
		}
	}

	basePk, err := solana.PublicKeyFromBase58(pair.BaseMint)
	if err != nil {
		return base, quote, err
	}

	quotePk, err := solana.PublicKeyFromBase58(pair.QuoteMint)
	if err != nil {
		return base, quote, err
	}

	base, ok = symbolMap[basePk.String()]
	if !ok {
		base = nameSplit[0]
	}

	quote, ok = symbolMap[quotePk.String()]
	if !ok {
		quote = nameSplit[1]
	}

	// protect against setting "" in the db
	if base == "" {
		base = symbols.TargetUnknown
	}

	if quote == "" {
		quote = symbols.TargetUnknown
	}

	return base, quote, nil
}

// chunkedRequests runs GetMultipleAccounts requests chunked and in parallel.  One GetMultipleAccounts request
// is limited to the chunkSize.
func (ig *Ingester) chunkedRequests(ctx context.Context, pairs Pairs, chunkSize int) ([]*rpc.Account, error) {
	totalAccounts := len(pairs)
	respAccounts := make([]*rpc.Account, totalAccounts)

	var wg sync.WaitGroup

	for i := 0; i < totalAccounts; i += chunkSize {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()

			end := start + chunkSize
			if end > totalAccounts {
				end = totalAccounts
			}

			reqAccounts := make([]solana.PublicKey, end-start)
			for j := range reqAccounts {
				pair := pairs[start+j]
				reqAccounts[j] = solana.MustPublicKeyFromBase58(pair.AmmID)
			}

			accountsResp, err := ig.client.GetMultipleAccounts(ctx, reqAccounts)
			if err != nil {
				ig.logger.Error("failed to query accounts", zap.Error(err))
				return
			}

			for j, account := range accountsResp {
				respAccounts[start+j] = account
			}
		}(i)
	}
	wg.Wait()

	return respAccounts, nil
}

var knownQuotes = []string{
	"SOL",
}
