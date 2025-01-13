package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd/utils/validate"
	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/lib/file"
	"github.com/skip-mev/connect-mmu/validator"
	validatortypes "github.com/skip-mev/connect-mmu/validator/types"
)

const (
	flagMarketmap       = "market-map"
	flagOutput          = "output"
	flagConnectVersion  = "connect-version"
	flagStartDelay      = "start-delay"
	flagDuration        = "duration"
	flagOracleConfig    = "oracle-config"
	flagWriteToStdError = "write-to-stderr"
	flagCMCAPIKey       = "cmc-api-key" //nolint:gosec
	flagHealthFile      = "health-file"
	flagEnableAll       = "enable-all"
	flagEnableMarkets   = "enable-markets"
	flagEnableOnly      = "enable-only"

	// stats flags
	flagZScoreBound             = "zscore-bound"
	flagReferencePriceAllowance = "reference-price-allowance"
	flagSuccessThreshold        = "success-threshold"

	cmcKeyEnvVar = "CMC_API_KEY"
)

// NOTE: This command requires you to have both `connect` and `validator` installed.
// To install validator run `make install-validator` in the root of this repo.
// To install connect run `make install` in the Connect repo.
func ValidateCmd() *cobra.Command {
	var flags validateCmdFlags

	cmd := &cobra.Command{
		Use:     "validate",
		Short:   "checks the healthiness of a marketmap in a real Connect instance",
		Long:    "ingests logs from Connect and outputs a healthcheck of all currency_pair/providers",
		Example: "validate --market-map marketmap.json --oracle-config oracle.json --start-delay 10s --duration 1m",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := checkInstalled("sentry"); err != nil {
				return err
			}

			if flags.successThreshold <= 0 {
				return errors.New("invalid success threshold")
			}

			mm, err := getMarketMapFromFlags(flags)
			if err != nil {
				return err
			}

			if err := validate.ApplyOptionsToMarketMap(mm, flags.enableAll, flags.enableOnly, flags.enableMarkets); err != nil {
				return err
			}

			cmd.Printf("running validation for %d markets\n", len(mm.Markets))

			cmcAPIKey := flags.cmcAPIKey
			if cmcAPIKey == "" {
				cmcAPIKey = os.Getenv(cmcKeyEnvVar)
			}
			if cmcAPIKey != "" {
				cmd.Println("reference price checking enabled")
			} else {
				cmd.Println("reference price checking disabled")
			}

			// write this new marketmap to a temp file, so we can pass the filepath to connect.
			f, err := writeMarketMapToTempFile(mm)
			if err != nil {
				return err
			}

			// connect command
			connectBin := "connect"
			if flags.connectVersion != "" {
				connectBin += "-" + flags.connectVersion
			}
			if err := checkInstalled(connectBin); err != nil {
				return err
			}

			// check version
			verOut, err := exec.Command(connectBin, "version").Output()
			if err != nil {
				return fmt.Errorf("failed to execute connect version command: %w", err)
			}

			cmd.Printf("using %s %s", connectBin, string(verOut))

			connectCmd := []string{connectBin, "--market-config-path", f.Name(), "--log-std-out-level", "debug"}
			if flags.oracleConfig != "" {
				_, err := os.Stat(flags.oracleConfig)
				if err != nil {
					return fmt.Errorf("failed to find oracle config file %s: %w", flags.oracleConfig, err)
				}

				connectCmd = append(connectCmd, "--oracle-config", flags.oracleConfig)
			}
			const (
				redirect = "2>&1"
				pipe     = "|"
			)

			// validator command
			ingestionRun := []string{
				"sentry",
				"run",
				fmt.Sprintf("--%s", flagStartDelay), flags.startDelay.String(), // start delay
				fmt.Sprintf("--%s", flagDuration), flags.duration.String(), // duration
				fmt.Sprintf("--%s", flagHealthFile), flags.healthFile,
			}
			// create the full command which is the connect command, pipe operator, then validator command.
			fullCmd := make([]string, 0, len(connectCmd))

			fullCmd = append(fullCmd, connectCmd...)   // connect
			fullCmd = append(fullCmd, redirect)        // redirect
			fullCmd = append(fullCmd, pipe)            // pipe
			fullCmd = append(fullCmd, ingestionRun...) // validator

			cmdString := strings.Join(fullCmd, " ")

			// we run the command with `sh -c` because os/exec by itself cannot handle multiple binaries in one exec.
			command := exec.Command("sh", "-c", cmdString)
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr

			// catch ctrl-c.
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT)

			// run the command.
			err = command.Run()
			if err != nil {
				return fmt.Errorf("error running command: %w", err)
			}

			// non-blocking channel read. the command will return when the user hits ctrl-c OR if the timer ends.
			// firstly, we need the ctrl-c handler here because we don't necessarily want _this_ program to exit
			// when that happens, just the child process. however, we don't want to block on that channel read either,
			// because the program could exit bc the duration ended.
			select {
			case <-sigs:
			default:
			}

			cmd.Println("validation finished")

			// read the health report file written by the validation's log ingestion binary.
			health, err := file.ReadJSONFromFile[validatortypes.MarketHealth](flags.healthFile)
			if err != nil {
				return fmt.Errorf("failed to get health file: %w", err)
			}

			// pass info to validator, generate reports.
			val := validator.New(mm, validator.WithCMCKey(cmcAPIKey))
			reports, err := val.Report(cmd.Context(), health)
			if err != nil {
				return fmt.Errorf("failed to generate report: %w", err)
			}

			// grade reports using the bounds. this will mark the reports as PASS or FAIL
			// as well as providing a top level success ratio of a market.
			summary := val.GradeReports(
				reports,
				validator.CheckZScore(flags.zScoreBound),
				validator.CheckSuccessThreshold(float64(flags.successThreshold)),
				validator.CheckReferencePrice(float64(flags.referencePriceAllowance)),
			)

			if err := file.WriteJSONToFile(flags.writeReportFile, summary); err != nil {
				cmd.Println("failed to write report file: ", err.Error())
			}

			allErrs := generateErrorFromReport(mm, summary, val.MissingReports(health))

			logger := logging.Logger(cmd.Context())
			logger.Info("validation errors", zap.Bool("slack_report", true), zap.Errors("errors", allErrs))

			err = errors.Join(allErrs...)
			return err
		},
	}

	validateCmdConfigureFlags(cmd, &flags)

	return cmd
}

// validateCmdFlags is a convenience container containing all flag values.
type validateCmdFlags struct {
	// connectVersion to use in validation. DOCKER ONLY.
	connectVersion string
	// the path to read a marketmap from. this will run connect with this marketmap during the validation.
	marketmapPath string
	// duration is the duration the validation will run.
	duration time.Duration
	// startDelay is how long the validation service will let connect run before ingesting logs.
	startDelay time.Duration
	// oracleConfig is the oracle config to pass to the connect instance. this is useful when providers require API keys.
	oracleConfig string
	// healthFile is the file path to output the health report to
	healthFile string
	// enableAll is an option to enable all markets before running validation.
	enableAll bool
	// enableMarkets is an option to enable the specified markets in the marketmap. does not disable others.
	enableMarkets []string
	// enableOnly is an option to ONLY enable the specified markets in the marketmap. does disable others.
	enableOnly []string

	// writeToStdErr will write the results to std error. this is useful for making sure the job fails a workflow pipeline.
	writeToStdErr bool

	// write file for the final reports.
	writeReportFile string

	// stats
	//
	// successThreshold is a percent value that will mark providers as failed if their success % is not >= the threshold.
	successThreshold int
	// zScoreBound is the bounds of the zScore. If a provider's zScore exceeds these bounds, the provider will be marked failed.
	zScoreBound float64
	// referencePriceAllowance is the allowed percent difference in reference price for a provider to still be considered valid.
	// if a provider's reference price difference exceeds this allowance, it will be marked failed.
	referencePriceAllowance int

	// cmcAPIKey is an optional api key that will be used by the validator to generate reference price percent differences.
	cmcAPIKey string
}

func validateCmdConfigureFlags(cmd *cobra.Command, flags *validateCmdFlags) {
	cmd.Flags().IntVar(&flags.successThreshold, flagSuccessThreshold, 60, "percentage value of when a market should no longer be considered healthy. (i.e. 50 would mean the provider needs a 50/50 success/failure ratio, 100 would mean no tolerance for failures at all)")
	cmd.Flags().DurationVar(&flags.startDelay, flagStartDelay, 1*time.Minute, "the amount of time the process will wait until it begins reading logs")
	cmd.Flags().DurationVar(&flags.duration, flagDuration, 5*time.Minute, "the amount of time the process will run before exiting")
	cmd.Flags().StringVar(&flags.marketmapPath, flagMarketmap, "", "optional path to marketmap file to output potential anomalies such as missing reports")
	cmd.Flags().StringVar(&flags.oracleConfig, flagOracleConfig, "", "use this flag to pass in an oracle config to connect. this is useful if your markets require API keys")
	cmd.Flags().BoolVar(&flags.writeToStdErr, flagWriteToStdError, true, "write the results as an error to std error")
	cmd.Flags().StringVar(&flags.connectVersion, flagConnectVersion, "", "DOCKER ONLY: the connect version to run the validation on. if empty, the latest will be used. examples: 1.0.12, 2.0.0")
	cmd.Flags().StringVar(&flags.cmcAPIKey, flagCMCAPIKey, "", "coinmarketcap API key that will be used to get reference prices to check provider prices against")
	cmd.Flags().Float64Var(&flags.zScoreBound, flagZScoreBound, 2.0, "bound for ZScore before considered unhealthy")
	cmd.Flags().IntVar(&flags.referencePriceAllowance, flagReferencePriceAllowance, 15, "percent reference price difference allowance")
	cmd.Flags().StringVar(&flags.healthFile, flagHealthFile, "health.json", "the output path for the health file")
	cmd.Flags().StringVar(&flags.writeReportFile, flagOutput, "reports.json", "the output path for the reports")
	cmd.Flags().BoolVar(&flags.enableAll, flagEnableAll, false, "enable all markets for validation")
	cmd.Flags().StringSliceVar(&flags.enableMarkets, flagEnableMarkets, nil, "enable the specified markets. NOTE: this will not disable markets that are currently enabled.")
	cmd.Flags().StringSliceVar(&flags.enableOnly, flagEnableOnly, nil, "enable ONLY the specified markets. all other markets will be disabled")

	cmd.MarkFlagRequired(flagMarketmap)
	cmd.MarkFlagsMutuallyExclusive(flagEnableMarkets, flagEnableOnly, flagEnableAll)
}

// generateErrorFromReport will generate an error based on failing and missing reports.
func generateErrorFromReport(mm types.MarketMap, reports validatortypes.Reports, missing map[string][]string) []error {
	allErrs := make([]error, 0)
	for ticker, providers := range missing {
		allErrs = append(allErrs,
			fmt.Errorf("missing %s from: %s", ticker, strings.Join(providers, ",")),
		)
	}
	reportErrs := make([]error, 0)
	for _, report := range reports.Reports {
		providerErrs := make([]error, 0)
		for _, providerReport := range report.ProviderReports {
			if providerReport.Grade == validatortypes.GradeFailed {
				refPriceDiff := 0.0
				if providerReport.ReferencePriceDiff != nil {
					refPriceDiff = *providerReport.ReferencePriceDiff
				}
				providerErrs = append(providerErrs, fmt.Errorf(
					"%s: SuccessRate %f, ZScore %f, AveragePrice %f, ReferencePriceDiff: %f",
					providerReport.Name, providerReport.SuccessRate, providerReport.ZScore, providerReport.AveragePrice, refPriceDiff,
				))
			}
		}
		if len(providerErrs) > 0 {
			refPrice := ""
			if report.ReferencePrice != nil {
				refPrice = fmt.Sprintf("($%f)", *report.ReferencePrice)
			}
			providerErr := errors.Join(providerErrs...)
			reportErrs = append(reportErrs,
				fmt.Errorf("%s %s has %d/%d failing providers: %w",
					report.Ticker,
					refPrice,
					len(providerErrs),
					len(mm.Markets[report.Ticker].ProviderConfigs),
					providerErr,
				),
			)
		}
	}
	allErrs = append(allErrs, reportErrs...)

	return allErrs
}

// getMarketMapFromFlags will get a marketmap based on the flags passed.
func getMarketMapFromFlags(opts validateCmdFlags) (types.MarketMap, error) {
	mm, err := file.ReadMarketMapFromFile(opts.marketmapPath)
	if err != nil {
		return types.MarketMap{}, fmt.Errorf("error loading marketmap: %w", err)
	}
	if len(mm.Markets) == 0 {
		return types.MarketMap{}, fmt.Errorf("empty marketmap")
	}

	return mm, nil
}

func checkInstalled(bin string) error {
	_, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("required binary %q is not installed: %w", bin, err)
	}
	return nil
}

func writeMarketMapToTempFile(mm types.MarketMap) (*os.File, error) {
	f, err := os.CreateTemp("", "all_enabled_marketmap.*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for marketmap: %w", err)
	}
	bz, err := json.Marshal(mm)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fetched marketmap: %w", err)
	}
	_, err = f.Write(bz)
	if err != nil {
		return nil, fmt.Errorf("error writing to %s: %w", f.Name(), err)
	}
	return f, nil
}
