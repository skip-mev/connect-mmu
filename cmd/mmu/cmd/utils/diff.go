package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/josephburnett/jd/v2"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types/tickermetadata"
	slinkymarketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/connect-mmu/client/marketmap"
	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/lib/file"
)

var netsToRPC = map[string]string{
	"dydx-testnet":    "dydx-testnet-grpc.polkachu.com:23890",
	"dydx-mainnet":    "dydx-grpc.polkachu.com:23890",
	"neutron-mainnet": "neutron-grpc.polkachu.com:19190",
	"neutron-testnet": "neutron-testnet-grpc.polkachu.com:19190",
	"warden-testnet":  "warden-testnet-grpc.polkachu.com:27390",
	"initia-testnet":  "initia-testnet-grpc.polkachu.com:25790",
}

func DiffCmd() *cobra.Command {
	var flags diffCmdFlags

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "generate a diff between a generated marketmap and an on-chain marketmap",
		Long: "diff a generated marketmap and a live, on-chain marketmap. the +/- symbols indicate changes that will" +
			" be made to the marketmap if the generated marketmap were to overwrite the chain's marketmap.",
		Example: "diff --market-map mm.json -n dydx-testnet",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var grpcURL string

			// getting the network API url.
			switch {
			case flags.networkName != "":
				grpcURL = netsToRPC[flags.networkName]
				if grpcURL == "" {
					return fmt.Errorf("unknown network pick from: %s", strings.Join(maps.Keys(netsToRPC), ","))
				}
			case flags.networkURL != "":
				grpcURL = flags.networkURL
			default:
				return errors.New("empty network")
			}

			chainMM, err := getMarketMap(cmd.Context(), grpcURL, flags.useSlinkyAPI)
			if err != nil {
				return err
			}

			generatedMarketMap, err := file.ReadMarketMapFromFile(flags.marketMapPath)
			if err != nil {
				return fmt.Errorf("unable to read file marketmap: %w", err)
			}

			// we don't have removals yet. so anything on-chain
			// that's _not_ in generated gets deleted, so it doesn't clutter the diff.
			for name := range chainMM.Markets {
				if _, ok := generatedMarketMap.Markets[name]; !ok {
					delete(chainMM.Markets, name)
				}
			}

			sortProviderConfigs(generatedMarketMap)
			sortProviderConfigs(chainMM)
			// if we're not showing ref price change, we mute the ref/liq changes by setting them to zero.
			if !flags.showRefPriceChanges {
				muteRefPriceAndVolumeChanges(generatedMarketMap)
				muteRefPriceAndVolumeChanges(chainMM)
			}

			// new markets should be more prominent. we don't want to mix them with the main changes diff.
			newMarkets := removeNewMarkets(generatedMarketMap, chainMM)

			// after everything has been modify, we marshal the mm's to json.
			chainMarketMapBz, err := json.Marshal(chainMM)
			if err != nil {
				return err
			}
			localMM, err := json.Marshal(generatedMarketMap)
			if err != nil {
				return err
			}

			// pass em into the differ
			generated, err := jd.ReadJsonString(string(localMM))
			if err != nil {
				return err
			}
			chainNode, err := jd.ReadJsonString(string(chainMarketMapBz))
			if err != nil {
				return err
			}

			// diff
			theDiff := chainNode.Diff(generated, jd.COLOR).Render(jd.COLOR)
			changes := strings.Count(theDiff, "@") // each diff is prefixed with an @
			fmt.Printf("\n\n=== %d CHANGES ===\n\n", changes)
			theDiff = strings.ReplaceAll(theDiff, "@", "\n@") // space em out.
			cmd.Println(theDiff)

			var newMarketsString string
			if len(newMarkets) > 0 {
				newMarketsTickers := make([]string, 0)

				colorDefault := "\033[0m"
				colorGreen := "\033[32m"

				b := bytes.NewBuffer(nil)
				b.WriteString(fmt.Sprintf("\n\n=== ADDING %d NEW MARKETS ===\n\n ", len(newMarkets)))
				for _, newMarket := range newMarkets {
					newMarketsTickers = append(newMarketsTickers, newMarket.Ticker.String())
					b.WriteString(colorGreen)
					b.WriteString("+ ")
					b.WriteString(newMarket.Ticker.String())
					b.WriteString("\n")
					b.WriteString(colorDefault)
					bz, err := json.Marshal(newMarket)
					if err != nil {
						panic(err)
					}
					b.WriteString(string(bz))
					b.WriteString("\n\n")
				}
				newMarketsString = b.String()
				fmt.Println(newMarketsString)

				logger := logging.Logger(cmd.Context())
				logger.Info("new markets", zap.Bool("slack_report", true), zap.Strings("markets", newMarketsTickers))
			} else {
				fmt.Printf("\n\n=== NO NEW MARKETS ADDED ===\n\n")
			}

			if flags.outputPath != "" {
				bz := bytes.NewBuffer([]byte(theDiff))
				bz.WriteString(newMarketsString)
				err := file.WriteBytesToFile(flags.outputPath+".txt", bz.Bytes())
				if err != nil {
					return fmt.Errorf("failed to write diff to file: %w", err)
				}
			}

			return nil
		},
	}

	diffCmdConfigureFlags(cmd, &flags)

	return cmd
}

type diffCmdFlags struct {
	outputPath          string
	marketMapPath       string
	showRefPriceChanges bool
	networkName         string
	networkURL          string
	useSlinkyAPI        bool
}

func diffCmdConfigureFlags(cmd *cobra.Command, flags *diffCmdFlags) {
	const (
		flagNetwork, flagNetworkShort       = "network", "n"
		flagNetworkRaw, flagNetworkRawShort = "network-raw", "r"
		flagShowReferencePriceChanges       = "show-reference-price"
		flagSlinkyAPI                       = "slinky-api"
	)

	cmd.Flags().StringVar(&flags.outputPath, flagOutput, "", "writes the diff to a file")
	cmd.Flags().StringVar(&flags.marketMapPath, flagMarketmap, "", "load a marketmap from a file")
	cmd.Flags().BoolVar(&flags.showRefPriceChanges, flagShowReferencePriceChanges, false, "show changes in reference price and liquidity")
	cmd.Flags().StringVarP(&flags.networkName, flagNetwork, flagNetworkShort, "", "blockchain network to query i.e. dydx-testnet")
	cmd.Flags().StringVarP(&flags.networkURL, flagNetworkRaw, flagNetworkRawShort, "", "raw blockchain gRPC network URL to query. i.e. dydx-testnet-grpc.polkachu.com")
	cmd.Flags().BoolVar(&flags.useSlinkyAPI, flagSlinkyAPI, false, "use the slinky API to query the marketmap")

	cmd.MarkFlagsOneRequired(flagNetwork, flagNetworkRaw)
	cmd.MarkFlagsMutuallyExclusive(flagNetwork, flagNetworkRaw)
}

// getMarketMap fetches the marketmap from the chain.
func getMarketMap(ctx context.Context, grpcURL string, useSlinky bool) (marketmaptypes.MarketMap, error) {
	c, err := grpc.NewClient(grpcURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return marketmaptypes.MarketMap{}, err
	}

	var client marketmap.Client
	if useSlinky {
		client = marketmap.NewSlinkyModuleMarketMapClient(slinkymarketmaptypes.NewQueryClient(c), zap.NewNop())
	} else {
		client = marketmap.NewConnectModuleMarketMapClient(marketmaptypes.NewQueryClient(c), zap.NewNop())
	}

	return client.GetMarketMap(ctx)
}

func sortProviderConfigs(mm marketmaptypes.MarketMap) {
	for name, market := range mm.Markets {
		providers := market.ProviderConfigs
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].Name < providers[j].Name
		})
		market.ProviderConfigs = providers
		mm.Markets[name] = market
	}
}

func muteRefPriceAndVolumeChanges(mm marketmaptypes.MarketMap) {
	for name, market := range mm.Markets {
		if market.Ticker.Metadata_JSON != "" {
			dydx, err := tickermetadata.DyDxFromJSONString(market.Ticker.Metadata_JSON)
			if err != nil {
				continue // lets just not fail.
			}
			dydx.ReferencePrice = 0
			dydx.Liquidity = 0
			dydxBz, err := json.Marshal(dydx)
			if err != nil {
				panic(err)
			}
			market.Ticker.Metadata_JSON = string(dydxBz)
			mm.Markets[name] = market
		}
	}
}

func removeNewMarkets(generated, onchain marketmaptypes.MarketMap) []marketmaptypes.Market {
	newMarkets := make([]marketmaptypes.Market, 0)
	for name, market := range generated.Markets {
		if _, ok := onchain.Markets[name]; !ok {
			newMarkets = append(newMarkets, market)
			delete(generated.Markets, name)
		}
	}
	return newMarkets
}
