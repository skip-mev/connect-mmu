package basic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	slinkymmtypes "github.com/skip-mev/slinky/x/marketmap/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/dispatcher"
	"github.com/skip-mev/connect-mmu/dispatcher/transaction/generator"
	"github.com/skip-mev/connect-mmu/lib/file"
	"github.com/skip-mev/connect-mmu/signing"
	"github.com/skip-mev/connect-mmu/signing/simulate"
)

// DispatchCmd returns a command to DispatchCmd market upserts.
func DispatchCmd(signingRegistry *signing.Registry) *cobra.Command {
	var flags dispatchCmdFlags

	cmd := &cobra.Command{
		Use:     "dispatch",
		Short:   "dispatch either upserts or removals to a chain",
		Example: "dispatch --config path/to/config.json --upserts path/to/upserts.json --simulate",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := logging.Logger(cmd.Context())

			// Validate that exactly one of upserts or removals is set
			hasUpserts := flags.upsertsPath != ""
			hasRemovals := flags.removalsPath != ""
			if !hasUpserts && !hasRemovals {
				return fmt.Errorf("must specify either --upserts and/or --removals")
			}

			cfg, err := config.ReadConfig(flags.configPath)
			if err != nil {
				logger.Error("failed to load config", zap.Error(err))
				return err
			}

			if cfg.Dispatch == nil {
				return errors.New("dispatch configuration missing from mmu config")
			}

			if cfg.Chain == nil {
				return errors.New("chain configuration missing from mmu config")
			}

			signerConfig := cfg.Dispatch.SigningConfig
			if flags.simulateAddress != "" {
				signerConfig = config.SigningConfig{
					Type:   simulate.TypeName,
					Config: simulate.SigningAgentConfig{Address: flags.simulateAddress},
				}
			}

			signer, err := signingRegistry.CreateSigner(signerConfig, *cfg.Chain)
			if err != nil {
				return fmt.Errorf("failed to create signer: %w", err)
			}

			dp, err := dispatcher.New(*cfg.Dispatch, *cfg.Chain, signer, logger)
			if err != nil {
				return fmt.Errorf("failed to create dispatcher: %w", err)
			}

			signerAddress, err := getSignerAddress(cmd.Context(), signer)
			if err != nil {
				return fmt.Errorf("failed to get signer address: %w", err)
			}

			var txs []cmttypes.Tx
			if hasUpserts {
				upsertTxs, err := generateUpsertTransactions(cmd.Context(), logger, dp, &cfg, signerAddress, flags.upsertsPath)
				if err != nil {
					return err
				}
				txs = append(txs, upsertTxs...)
			}
			if hasRemovals {
				removalTxs, err := generateRemovalTransactions(cmd.Context(), logger, dp, &cfg, signerAddress, flags.removalsPath)
				if err != nil {
					return err
				}
				txs = append(txs, removalTxs...)
			}

			decodedTxs, err := decodeTxs(txs)
			if err != nil {
				return err
			}

			err = file.WriteJSONToFile("transactions.json", decodedTxs)
			if err != nil {
				return err
			}

			if flags.simulate {
				return nil
			}

			return dp.SubmitTransactions(cmd.Context(), txs)
		},
	}

	dispatchCmdConfigureFlags(cmd, &flags)

	return cmd
}

func generateUpsertTransactions(
	ctx context.Context,
	logger *zap.Logger,
	dp *dispatcher.Dispatcher,
	cfg *config.Config,
	signerAddress string,
	upsertsPath string,
) ([]cmttypes.Tx, error) {
	upserts, err := file.ReadJSONFromFile[[]mmtypes.Market](upsertsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read upserts file: %w", err)
	}

	upsertMsgs, err := generator.ConvertUpsertsToMessages(
		logger,
		cfg.Dispatch.TxConfig,
		cfg.Chain.Version,
		signerAddress,
		upserts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert upserts to messages: %w", err)
	}

	txs, err := dp.GenerateTransactions(ctx, upsertMsgs)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func generateRemovalTransactions(
	ctx context.Context,
	logger *zap.Logger,
	dp *dispatcher.Dispatcher,
	cfg *config.Config,
	signerAddress string,
	removalsPath string,
) ([]cmttypes.Tx, error) {
	removals, err := file.ReadJSONFromFile[[]string](removalsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read marketmap removals file: %w", err)
	}

	removalMsgs, err := generator.ConvertRemovalsToMessages(
		logger,
		cfg.Chain.Version,
		signerAddress,
		removals,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert removals to messages: %w", err)
	}

	txs, err := dp.GenerateTransactions(ctx, removalMsgs)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

type DecodedTx struct {
	Body       *codectypes.Any `json:"body"`
	AuthInfo   *codectypes.Any `json:"auth_info"`
	Signatures [][]byte        `json:"signatures"`
}

func decodeTxs(txs []cmttypes.Tx) ([]DecodedTx, error) {
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	decoder := auth.DefaultTxDecoder(cdc)
	jsonEncoder := auth.DefaultJSONTxEncoder(cdc)
	slinkymmtypes.RegisterInterfaces(registry)

	decodedTxs := make([]DecodedTx, len(txs))

	for i, tx := range txs {
		decodedTx, err := decoder(tx)
		if err != nil {
			return nil, err
		}

		jsonBz, err := jsonEncoder(decodedTx)
		if err != nil {
			return nil, err
		}

		var decodedJSON DecodedTx

		if err := json.Unmarshal(jsonBz, &decodedJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tx json: %w", err)
		}

		decodedTxs[i] = decodedJSON
	}

	return decodedTxs, nil
}

func getSignerAddress(ctx context.Context, signer signing.SigningAgent) (string, error) {
	acc, err := signer.GetSigningAccount(ctx)
	if err != nil {
		return "", err
	}

	// convert to a base account
	baseAcc, ok := acc.(*authtypes.BaseAccount)
	if !ok {
		return "", fmt.Errorf("expected BaseAccount but got %T", acc)
	}
	return baseAcc.Address, nil
}

type dispatchCmdFlags struct {
	configPath      string
	upsertsPath     string
	removalsPath    string
	simulate        bool
	simulateAddress string
}

func dispatchCmdConfigureFlags(cmd *cobra.Command, flags *dispatchCmdFlags) {
	cmd.Flags().StringVar(&flags.configPath, ConfigPathFlag, ConfigPathDefault, ConfigPathDescription)
	cmd.Flags().StringVar(&flags.upsertsPath, UpsertsPathFlag, "", UpsertsPathDescription)
	cmd.Flags().StringVar(&flags.removalsPath, RemovalsPathFlag, "", RemovalsPathDescription)
	cmd.Flags().BoolVar(&flags.simulate, SimulateFlag, SimulateDefault, SimulateDescription)
	cmd.Flags().StringVar(&flags.simulateAddress, SimulateAddressFlag, SimulateAddressDefault, SimulateAddressDescription)
}
