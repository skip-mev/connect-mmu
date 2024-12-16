//go:build test
// +build test

package test

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	cmthttp "github.com/cometbft/cometbft/rpc/client/http"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/skip-mev/connect-mmu/client/dydx"
	"github.com/skip-mev/connect-mmu/client/marketmap"
	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/dispatcher"
	txgenerator "github.com/skip-mev/connect-mmu/dispatcher/transaction/generator"
	"github.com/skip-mev/connect-mmu/dispatcher/transaction/submitter"
	"github.com/skip-mev/connect-mmu/generator"
	"github.com/skip-mev/connect-mmu/lib/grpc"
	indexer "github.com/skip-mev/connect-mmu/market-indexer"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/binance"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bybit"
	"github.com/skip-mev/connect-mmu/override"
	"github.com/skip-mev/connect-mmu/override/update"
	"github.com/skip-mev/connect-mmu/signing/local"
	"github.com/skip-mev/connect-mmu/store/provider"
	"github.com/skip-mev/connect-mmu/upsert"
)

const (
	dydxNodeGRPC = "localhost:9090"
	dydxNodeRPC  = "http://localhost:26657"
	dydxNodeRest = "http://localhost:1317"

	localMarketConfigPath     = "../local/fixtures/e2e/market.json"
	localGenerationConfigPath = "../local/fixtures/e2e/gen_dydx_localnet.json"
	localChainConfigPath      = "../local/fixtures/e2e/chain.json"
	localUpsertConfigPath     = "../local/fixtures/e2e/upsert.json"
	localSignerConfigPath     = "../local/fixtures/e2e/signer.json"
)

var (
	localMarketConfig     config.MarketConfig
	localGenerationConfig config.GenerateConfig
	localChainConfig      config.ChainConfig
	localUpsertConfig     config.UpsertConfig
	localSignerConfig     config.SigningConfig
)

// read in config files to use from our local test files
func initConfigurations() {
	bz, err := os.ReadFile(localMarketConfigPath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bz, &localMarketConfig); err != nil {
		panic(err)
	}

	bz, err = os.ReadFile(localGenerationConfigPath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bz, &localGenerationConfig); err != nil {
		panic(err)
	}

	bz, err = os.ReadFile(localChainConfigPath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bz, &localChainConfig); err != nil {
		panic(err)
	}

	bz, err = os.ReadFile(localUpsertConfigPath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bz, &localUpsertConfig); err != nil {
		panic(err)
	}

	bz, err = os.ReadFile(localSignerConfigPath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(bz, &localSignerConfig); err != nil {
		panic(err)
	}

	// set the addresses to use locally running dydx chain
	localChainConfig.Version = config.VersionConnect

	// update endpoints for chain config to use local
	localChainConfig.RPCAddress = dydxNodeRPC
	localChainConfig.GRPCAddress = dydxNodeGRPC
	localChainConfig.RESTAddress = dydxNodeRest

	// remove binance and bybit because they have geo restrictions
	delete(localGenerationConfig.Providers, binance.ProviderName)
	delete(localGenerationConfig.Providers, bybit.ProviderName)
	updatedIngesters := make([]config.IngesterConfig, 0)
	for _, ingester := range localMarketConfig.Ingesters {
		if ingester.Name != binance.Name && ingester.Name != bybit.Name {
			updatedIngesters = append(updatedIngesters, ingester)
		}
	}

	localMarketConfig.Ingesters = updatedIngesters
}

type E2ESuite struct {
	suite.Suite

	logger *zap.Logger

	idx            *indexer.Indexer
	gen            generator.Generator
	marketOverride override.MarketMapOverride
	dispatch       *dispatcher.Dispatcher
	mmClient       marketmap.Client
	dispatcherCfg  config.DispatchConfig
}

func (s *E2ESuite) SetupSuite() {
	initConfigurations()
	s.logger = zaptest.NewLogger(s.T())
	providerStore, err := provider.NewMemoryStoreFromFile(localMarketConfigPath)
	require.NoError(s.T(), err)

	conn, err := grpc.NewChainGrpcImpl(dydxNodeGRPC, false)
	require.NoError(s.T(), err)
	s.mmClient = marketmap.NewConnectModuleMarketMapClient(mmtypes.NewQueryClient(conn.ClientConn), s.logger)

	// create an info logger because indexer debug logs slow down exec
	infoLogger := zaptest.NewLogger(s.T(), zaptest.Level(zapcore.InfoLevel))

	// indexer
	s.idx, err = indexer.NewIndexer(localMarketConfig, infoLogger, providerStore)
	s.Require().NoError(err)

	// generator
	s.gen = generator.New(s.logger, providerStore)

	s.marketOverride, err = override.NewDyDxOverride(dydx.NewHTTPClient(dydxNodeRest))
	s.Require().NoError(err)

	// dispatcher
	s.dispatch = s.setupDispatcher()
}

func (s *E2ESuite) TestIndexGenerateDispatch() {
	ctx := context.Background()

	err := s.idx.Index(ctx)
	s.Require().NoError(err, "failed to index markets")

	generatedMarketMap, _, err := s.gen.GenerateMarketMap(ctx, localGenerationConfig)
	s.Require().NoError(err, "failed to generate marketmap")

	onChainMM, err := s.mmClient.GetMarketMap(ctx)
	require.NoError(s.T(), err, "failed to get marketmap")

	overriddenMarketMap, err := s.marketOverride.OverrideGeneratedMarkets(
		ctx,
		s.logger,
		onChainMM,
		generatedMarketMap,
		update.Options{
			UpdateEnabled:      false,
			OverwriteProviders: true,
			ExistingOnly:       false,
		},
	)
	s.Require().NoError(err, "failed to override generated market map")

	upserter, err := upsert.New(s.logger, localUpsertConfig, overriddenMarketMap, onChainMM)
	s.Require().NoError(err)
	upserts, err := upserter.GenerateUpserts()
	require.NoError(s.T(), err, "failed to generate upserts")
	bz, err := json.MarshalIndent(upserts, "", "  ")
	s.Require().NoError(err)

	err = os.WriteFile("upserts.json", bz, 0o600)
	require.NoError(s.T(), err)

	msgs, err := txgenerator.ConvertUpsertsToMessages(s.logger, s.dispatcherCfg.TxConfig, localChainConfig.Version, upserts)
	require.NoError(s.T(), err, "failed to convert upserts to messages")

	txs, err := s.dispatch.GenerateTransactions(ctx, msgs)
	require.NoError(s.T(), err, "failed to generate transactions")

	err = s.dispatch.SubmitTransactions(ctx, txs)
	require.NoError(s.T(), err, "failed to submit transactions")

	time.Sleep(5 * time.Second)
	s.checkOnChainMarkets(ctx, generatedMarketMap)
}

func (s *E2ESuite) checkOnChainMarkets(ctx context.Context, desired mmtypes.MarketMap) {
	// get mm after dispatch
	mmAfter, err := s.mmClient.GetMarketMap(ctx)
	s.Require().NoError(err)

	// resulting market map should have all the generated markets.
	for _, desiredMarket := range desired.Markets {
		onChainMarket, ok := mmAfter.Markets[desiredMarket.Ticker.String()]
		s.Require().True(ok, "resulting market map did NOT contain desired market %s: %v", desiredMarket.Ticker.String(), desiredMarket)

		// check that we updated all disabled markets to what we desired
		if !onChainMarket.Ticker.Enabled {
			s.Require().Equal(desiredMarket, onChainMarket)
		}
	}
}

func (s *E2ESuite) setupDispatcher() *dispatcher.Dispatcher {
	rpcClient, err := cmthttp.New(dydxNodeRPC, "")
	s.Require().NoError(err)
	txSubmitter := submitter.NewTransactionSubmitter(rpcClient, config.DefaultSubmitterConfig(), s.logger)

	s.dispatcherCfg = config.DispatchConfig{
		TxConfig: config.TransactionConfig{
			MaxBytesPerTx: math.MaxInt,
			MaxGas:        math.MaxInt,
			GasAdjustment: 1.5,
			MinGasPrice:   sdk.NewInt64DecCoin("adv4tnt", 25000000000),
		},
		SigningConfig: localSignerConfig,
	}

	chainCfg := config.ChainConfig{
		RPCAddress:  dydxNodeRPC,
		GRPCAddress: dydxNodeGRPC,
		RESTAddress: dydxNodeRest,
		DYDX:        true,
		Version:     config.VersionConnect,
		ChainID:     "localdydxprotocol",
		Prefix:      "dydx",
	}

	signer, err := local.NewSigningAgent(s.dispatcherCfg.SigningConfig.Config, chainCfg)
	s.Require().NoError(err)

	txProvider, err := txgenerator.NewSigningTransactionGeneratorFromConfig(
		s.dispatcherCfg,
		chainCfg,
		signer,
		s.logger,
	)
	s.Require().NoError(err)

	return dispatcher.NewFromClients(
		txProvider,
		txSubmitter,
		s.logger,
		s.dispatcherCfg,
	)
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}
