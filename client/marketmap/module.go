package marketmap

import (
	"context"
	"fmt"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	slinkymmtypes "github.com/skip-mev/slinky/x/marketmap/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/connect-mmu/config"
)

var (
	_ Client = &SlinkyModuleMarketMapClient{}
	_ Client = &ConnectModuleMarketMapClient{}
)

func NewClientFromChainConfig(logger *zap.Logger, cfg config.ChainConfig) (Client, error) {
	cc, err := grpc.NewClient(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to create chain grpc client", zap.Error(err))
		return nil, err
	}

	var client Client
	switch cfg.Version {
	case config.VersionSlinky:
		client = NewSlinkyModuleMarketMapClient(slinkymmtypes.NewQueryClient(cc), logger)
	case config.VersionConnect:
		client = NewConnectModuleMarketMapClient(mmtypes.NewQueryClient(cc), logger)
	default:
		return nil, fmt.Errorf("unsupported chain version: %s", cfg.Version)
	}

	return client, nil
}

// SlinkyModuleMarketMapClient is a market-map provider that is capable of fetching market-maps from
// the x/marketmap module.
type SlinkyModuleMarketMapClient struct {
	marketMapModuleClient slinkymmtypes.QueryClient
	logger                *zap.Logger
}

// NewSlinkyModuleMarketMapClient creates a new SlinkyModuleMarketMapClient.
func NewSlinkyModuleMarketMapClient(marketMapModuleClient slinkymmtypes.QueryClient, logger *zap.Logger) *SlinkyModuleMarketMapClient {
	return &SlinkyModuleMarketMapClient{
		marketMapModuleClient: marketMapModuleClient,
		logger:                logger,
	}
}

// GetMarketMap retrieves a market-map from the x/marketmap module.
func (s *SlinkyModuleMarketMapClient) GetMarketMap(ctx context.Context) (mmtypes.MarketMap, error) {
	// get the market-map from x/marketmap
	// TODO(nikhil): consider handling last-updated here
	mm, err := s.marketMapModuleClient.MarketMap(ctx, &slinkymmtypes.MarketMapRequest{})
	if err != nil {
		s.logger.Error("error fetching market-map from slinky x/marketmap", zap.Error(err))
		return mmtypes.MarketMap{}, fmt.Errorf("error fetching market-map from slinky x/marketmap: %w", err)
	}

	// if entry is nil, return an empty market-map
	if mm.MarketMap.Markets == nil {
		mm.MarketMap.Markets = make(map[string]slinkymmtypes.Market)
	}

	// now convert to connect type
	connectMM := SlinkyToConnectMarketMap(mm.MarketMap)

	return connectMM, nil
}

// ConnectModuleMarketMapClient is a market-map provider that is capable of fetching market-maps from
// the x/marketmap module.
type ConnectModuleMarketMapClient struct {
	marketMapModuleClient mmtypes.QueryClient
	logger                *zap.Logger
}

// NewConnectModuleMarketMapClient creates a new ConnectModuleMarketMapClient.
func NewConnectModuleMarketMapClient(marketMapModuleClient mmtypes.QueryClient, logger *zap.Logger) *ConnectModuleMarketMapClient {
	return &ConnectModuleMarketMapClient{
		marketMapModuleClient: marketMapModuleClient,
		logger:                logger,
	}
}

// GetMarketMap retrieves a market-map from the x/marketmap module.
func (s *ConnectModuleMarketMapClient) GetMarketMap(ctx context.Context) (mmtypes.MarketMap, error) {
	// get the market-map from x/marketmap
	// TODO(nikhil): consider handling last-updated here
	mm, err := s.marketMapModuleClient.MarketMap(ctx, &mmtypes.MarketMapRequest{})
	if err != nil {
		s.logger.Error("error fetching market-map from connect x/marketmap", zap.Error(err))
		return mmtypes.MarketMap{}, fmt.Errorf("error fetching market-map from connect x/marketmap: %w", err)
	}

	// if entry is nil, return an empty market-map
	if mm.MarketMap.Markets == nil {
		mm.MarketMap.Markets = make(map[string]mmtypes.Market)
	}

	return mm.MarketMap, nil
}
