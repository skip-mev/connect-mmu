package basic

import (
	"errors"
	"fmt"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
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
func DispatchCmd(registry *signing.Registry) *cobra.Command {
	var flags dispatchCmdFlags

	cmd := &cobra.Command{
		Use:     "dispatch",
		Short:   "dispatch a upserts to a chain",
		Example: "dispatch --config path/to/config.json --upserts path/to/upserts.json --simulate",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := logging.Logger(cmd.Context())

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

			upserts, err := file.ReadJSONFromFile[[]mmtypes.Market](flags.upsertsPath)
			if err != nil {
				return fmt.Errorf("failed to read upserts file: %w", err)
			}

			msgs, err := generator.ConvertUpsertsToMessages(logger, cfg.Dispatch.TxConfig, cfg.Chain.Version, upserts)
			if err != nil {
				return fmt.Errorf("failed to convert upserts to messages: %w", err)
			}

			logger.Info("creating signer", zap.String("signer_type", cfg.Dispatch.SigningConfig.Type))

			signerConfig := cfg.Dispatch.SigningConfig
			if flags.simulateAddress != "" {
				signerConfig = config.SigningConfig{
					Type:   simulate.TypeName,
					Config: simulate.SigningAgentConfig{Address: flags.simulateAddress},
				}
			}

			signer, err := registry.CreateSigner(signerConfig, *cfg.Chain)
			if err != nil {
				return fmt.Errorf("failed to create signer: %w", err)
			}

			dp, err := dispatcher.New(*cfg.Dispatch, *cfg.Chain, signer, logger)
			if err != nil {
				return fmt.Errorf("failed to create dispatcher: %w", err)
			}

			txs, err := dp.GenerateTransactions(cmd.Context(), msgs)
			if err != nil {
				return err
			}

			err = file.WriteJSONToFile("transactions.json", txs)
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

type dispatchCmdFlags struct {
	configPath      string
	upsertsPath     string
	simulate        bool
	simulateAddress string
}

func dispatchCmdConfigureFlags(cmd *cobra.Command, flags *dispatchCmdFlags) {
	cmd.Flags().StringVar(&flags.configPath, ConfigPathFlag, ConfigPathDefault, ConfigPathDescription)
	cmd.Flags().StringVar(&flags.upsertsPath, UpsertsPathFlag, UpsertsPathDefault, UpsertsPathDescription)
	cmd.Flags().BoolVar(&flags.simulate, SimulateFlag, SimulateDefault, SimulateDescription)
	cmd.Flags().StringVar(&flags.simulateAddress, SimulateAddressFlag, SimulateAddressDefault, SimulateAddressDescription)
}
