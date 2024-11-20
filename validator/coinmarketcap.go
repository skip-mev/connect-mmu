package validator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types/tickermetadata"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	mapslib "github.com/skip-mev/connect-mmu/lib/maps"
	"github.com/skip-mev/connect-mmu/lib/slices"
	"github.com/skip-mev/connect-mmu/market-indexer/coinmarketcap"
)

var ErrCMCIDNotFound = errors.New("cmc ID not found")

// getReferencePrices gets reference prices from coinmarketcap.
func getReferencePrices(ctx context.Context, cmcAPIKey string, ids map[string]int64) (map[string]float64, error) {
	client := coinmarketcap.NewHTTPClient(cmcAPIKey)
	prices := sync.Map{}
	eg, ctx := errgroup.WithContext(ctx)

	cmcIDToTickers := mapslib.Invert(ids)

	reqSize := 1000
	cmcIDs := maps.Values(ids)
	chunks := slices.Chunk(cmcIDs, reqSize)

	for _, cmcChunk := range chunks {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				res, err := client.Quotes(ctx, cmcChunk)
				if err != nil {
					fmt.Printf("failed to fetch quotes: %v\n", err)
					return nil
				}
				for _, cmcID := range cmcChunk {
					data, ok := res.Data[fmt.Sprintf("%d", cmcID)]
					if !ok {
						fmt.Printf("no quote data for cmc ID: %d\n", cmcID)
						continue
					}

					priceData, ok := data.Quote["USD"]
					if !ok {
						return fmt.Errorf("price data not found for id %d: %w", cmcID, err)
					}
					for _, tickers := range cmcIDToTickers[cmcID] {
						prices.Store(tickers, priceData.Price)
					}
				}

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get reference prices: %w", err)
	}

	// Convert the sync map into a regular map.
	pricesMap := make(map[string]float64)
	prices.Range(func(k, v interface{}) bool {
		pricesMap[k.(string)] = v.(float64)
		return true
	})

	return pricesMap, nil
}

// getCMCIDMapping gets a mapping of ticker to cmc ID's for every market that has them.
func getCMCIDMapping(mm mmtypes.MarketMap) (map[string]int64, error) {
	cmcMapping := make(map[string]int64)
	for ticker, market := range mm.Markets {
		if market.Ticker.Metadata_JSON == "" {
			continue
		}
		var md tickermetadata.CoreMetadata
		err := json.Unmarshal([]byte(market.Ticker.Metadata_JSON), &md)
		if err != nil {
			// incorrectly formed metadata. skip.
			continue
		}
		id, err := getCMCIDFromMetadata(md)
		if err != nil { // only return the error if it wasn't an ID not found error.
			if errors.Is(err, ErrCMCIDNotFound) {
				continue
			}
			return nil, err
		}
		cmcMapping[ticker] = int64(id)
	}
	return cmcMapping, nil
}

// getCMCIDFromMetadata extracts the coinmarketcap ID from metadata.
// it will return an error if it couldn't convert the ID to an int,
// or if there was no coinmarketcap venue found in the metadata.
func getCMCIDFromMetadata(md tickermetadata.CoreMetadata) (int, error) {
	venue := "coinmarketcap"
	for _, id := range md.AggregateIDs {
		if id.Venue == venue {
			cmcID, err := strconv.Atoi(id.ID)
			if err != nil {
				return 0, err
			}
			return cmcID, nil
		}
	}
	return 0, ErrCMCIDNotFound
}
