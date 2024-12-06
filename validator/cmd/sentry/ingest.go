package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/skip-mev/connect-mmu/lib/file"
	"github.com/skip-mev/connect-mmu/validator/types"
)

type cmdFlags struct {
	healthFile string
	duration   time.Duration
	startDelay time.Duration
}

const (
	flagHealthFile = "health-file"
	flagStartDelay = "start-delay"
	flagDuration   = "duration"
)

// LogEntry represents the structure of Connect logs.
type LogEntry struct {
	Level        string `json:"level"`
	Timestamp    string `json:"ts"`
	Caller       string `json:"caller"`
	Message      string `json:"msg"`
	PID          int    `json:"pid"`
	Process      string `json:"process"`
	TargetTicker string `json:"target_ticker"`
	Provider     string `json:"provider"`
	Price        string `json:"price"`
	Error        string `json:"error"`
}

// connectAggregatorProcess is the process these logs are found in.
const connectAggregatorProcess = "index_price_aggregator"

// runCommand is a command that ingests connect logs and reports results of success/failures of market providers.
func runCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "ingests logs from Connect and outputs a report of all currency_pair/providers.",
		Long:    "ingests the logs of a Connect instance, reporting the number of successes, failures, and a rolling average price for every provider.",
		Example: "run --start-delay 10s --duration 1m",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := getFlags(cmd.Flags())
			cmd.Printf("watching logs for %s after %s\n", opts.duration.String(), opts.startDelay.String())
			cmd.Println("press ctrl-c to end the sentry run early")

			// counts = map[currency_pair]->map[provider]->Counts.
			counts := make(types.MarketHealth)

			// catch ctrl-c
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT)

			// the done channel will catch ctrl-c or the validation timer ending.
			done := make(chan bool, 1)

			// timer for how long the validation will run.
			timer := time.NewTimer(opts.duration)

			// startProcessing is signaled when the start-delay ends.
			startProcessing := make(chan bool, 1)

			// sleep for delay, then signal that we can process.
			go func() {
				time.Sleep(opts.startDelay)
				startProcessing <- true
			}()

			// handles ctrl-c signal and duration timer.
			go func() {
				select {
				case <-sigs:
					done <- true
				case <-timer.C:
					done <- true
				}
			}()

			// stdin reader
			reader := bufio.NewReader(os.Stdin)

			// processing will be set to true once the delay period has ended.
			processing := false

			// main logic handler.
			for {
				// select branch 1 -> we check if we're done, if not, process another line.
				select {
				case <-done: // done signal catches ctrl-c or duration ending. it will finalize the results of the validation.
					return finalize(counts, opts.healthFile)
				default:
					select {
					case <-startProcessing:
						processing = true
						cmd.Println("wait time over. processing logs now")
					default:
						// we don't do anything here. this case is so that it falls through to the next code block.
					}

					// advance the reader.
					line, err := reader.ReadString('\n')
					if err != nil {
						continue
					}

					// we do this here because we want the reader.ReadString to advance the lines. if we put this before,
					// the start-delay would be effectively useless.
					if !processing {
						continue
					}

					var logEntry LogEntry
					err = json.Unmarshal([]byte(line), &logEntry)
					if err != nil {
						continue
					}

					if logEntry.Process != connectAggregatorProcess {
						continue // We only care about the index_price_aggregator process logs
					}

					provider := logEntry.Provider
					targetTicker := logEntry.TargetTicker

					// this isn't a log we care about if these fields aren't populated.
					if provider == "" || targetTicker == "" {
						continue
					}

					// check if there's existing data for this ticker
					if counts[targetTicker] == nil {
						counts[targetTicker] = make(map[string]*types.Counts)
					}
					// check if we have set counts already for this ticker-provider pair.
					if counts[targetTicker][provider] == nil {
						counts[targetTicker][provider] = &types.Counts{}
					}

					// when logs include a price, that means the provider is healthy.
					if logEntry.Price != "" {
						price, err := strconv.ParseFloat(logEntry.Price, 64)
						if err != nil {
							continue
						}
						count := counts[targetTicker][provider]
						prevAvg := count.AveragePrice
						count.Success++
						count.AveragePrice = prevAvg + (price-prevAvg)/float64(count.Success)
						counts[targetTicker][provider] = count
					} else if logEntry.Error != "" {
						counts[targetTicker][provider].Failure++
					}
				}
			}
		},
	}

	cmd.Flags().Duration(flagStartDelay, 1*time.Minute, "the amount of time the process will wait until it begins reading logs")
	cmd.Flags().Duration(flagDuration, 10*time.Minute, "the amount of time the process will run before exiting")
	cmd.Flags().String(flagHealthFile, "health.json", "path to write the health report to")

	return cmd
}

func getFlags(flags *pflag.FlagSet) cmdFlags {
	duration, _ := flags.GetDuration(flagDuration)
	delay, _ := flags.GetDuration(flagStartDelay)
	healthFile, _ := flags.GetString(flagHealthFile)
	return cmdFlags{
		duration:   duration,
		startDelay: delay,
		healthFile: healthFile,
	}
}

// finalize calculates the missing reports and providers that did not meet the success threshold, then writes them to disk.
func finalize(health types.MarketHealth, healthFile string) error {
	err := file.CreateAndWriteJSONToFile(healthFile, health)
	if err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}
	return nil
}
