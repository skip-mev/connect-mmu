package generator

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	slinkymmtypes "github.com/skip-mev/slinky/x/marketmap/types"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/client/marketmap"
	"github.com/skip-mev/connect-mmu/config"
)

// ConvertUpsertsToMessages converts a set of upsert markets to a slice of sdk.Messages respecting the configured
// max size of a transaction.
func ConvertUpsertsToMessages(
	logger *zap.Logger,
	cfg config.TransactionConfig,
	version config.Version,
	authorityAddress string,
	upserts []mmtypes.Market,
) ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, 0)

	// create the update txs, such that the size of all markets per tx is optimized, while
	// not exceeding the max tx size
	currentTxSize := 0
	start := 0
	for i, market := range upserts {
		// fail if the market is invalid
		if err := market.ValidateBasic(); err != nil {
			logger.Error("invalid market", zap.Error(err))
			return nil, fmt.Errorf("invalid market: %w", err)
		}

		// validity check for market
		if market.Size() > cfg.MaxBytesPerTx {
			// if the market size exceeds the max tx size, then we can't create a tx for it (fail)
			logger.Error("market size exceeds max tx size", zap.Any("market", market), zap.Int("size",
				market.Size()), zap.Int("max_size", cfg.MaxBytesPerTx))
			return nil, fmt.Errorf("market size exceeds max tx size: %d > %d", market.Size(), cfg.MaxBytesPerTx)
		}

		// update the currentTxSize
		if currentTxSize+market.Size() > cfg.MaxBytesPerTx {
			// create the tx
			txMarkets := upserts[start:i]
			logger.Info("creating update msg", zap.Int("markets", len(txMarkets)))

			var msg sdk.Msg
			switch version {
			case config.VersionSlinky:
				msg = &slinkymmtypes.MsgUpsertMarkets{
					Authority: authorityAddress,
					Markets:   marketmap.ConnectToSlinkyMarkets(upserts),
				}
			case config.VersionConnect:
				msg = &mmtypes.MsgUpsertMarkets{
					Authority: authorityAddress,
					Markets:   upserts,
				}
			default:
				return nil, fmt.Errorf("unsupported version %s", version)
			}

			msgs = append(msgs, msg)

			// reset the currentTxSize
			currentTxSize = 0
			start = i
		}

		// add to the current group
		currentTxSize += market.Size()
	}

	// create the last tx
	if currentTxSize > 0 {
		var msg sdk.Msg
		switch version {
		case config.VersionSlinky:
			msg = &slinkymmtypes.MsgUpsertMarkets{
				Authority: authorityAddress,
				Markets:   marketmap.ConnectToSlinkyMarkets(upserts[start:]),
			}
		case config.VersionConnect:
			msg = &mmtypes.MsgUpsertMarkets{
				Authority: authorityAddress,
				Markets:   upserts[start:],
			}
		default:
			return nil, fmt.Errorf("unsupported version %s", version)
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// ConvertRemovalsToMessage converts a set of market tickers to remove to a slice of sdk.Message.
func ConvertRemovalsToMessages(
	logger *zap.Logger,
	version config.Version,
	authorityAddress string,
	removals []string,
) ([]sdk.Msg, error) {
	var msg sdk.Msg
	switch version {
	case config.VersionSlinky:
		msg = &slinkymmtypes.MsgRemoveMarkets{
			Authority: authorityAddress,
			Markets:   removals,
		}
	case config.VersionConnect:
		msg = &mmtypes.MsgRemoveMarkets{
			Authority: authorityAddress,
			Markets:   removals,
		}
	default:
		return nil, fmt.Errorf("unsupported version %s", version)
	}
	logger.Info("created remove msg", zap.Int("num markets to remove", len(removals)))
	return []sdk.Msg{msg}, nil
}
