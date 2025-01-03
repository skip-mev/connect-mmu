package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd"
	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/lib/aws"
	"github.com/skip-mev/connect-mmu/signing"
	"github.com/skip-mev/connect-mmu/signing/local"
	"github.com/skip-mev/connect-mmu/signing/simulate"

	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

type LambdaEvent struct {
	Command   string `json:"command"`
	Timestamp string `json:"timestamp,omitempty"`
	Network   string `json:"network,omitempty"`
}

type LambdaResponse struct {
	Timestamp string `json:"timestamp"`
}

type Command int

const (
	Index Command = iota
	Generate
	Override
	Validate
	Upserts
	Diff
)

func createSigningRegistry() *signing.Registry {
	r := signing.NewRegistry()
	err := errors.Join(
		r.RegisterSigner(simulate.TypeName, simulate.NewSigningAgent),
		r.RegisterSigner(local.TypeName, local.NewSigningAgent),
	)
	if err != nil {
		panic(err)
	}
	return r
}

func getSupportedCommands() map[string]Command {
	return map[string]Command{
		"index":    Index,
		"generate": Generate,
		"override": Override,
		"validate": Validate,
		"upserts":  Upserts,
		"diff":     Diff,
	}
}

func getSupportedNetworks() []string {
	return []string{"testnet", "mainnet"}
}

func getArgsFromLambdaEvent(ctx context.Context, event json.RawMessage) ([]string, error) {
	logger := logging.Logger(ctx)

	var lambdaEvent LambdaEvent
	if err := json.Unmarshal(event, &lambdaEvent); err != nil {
		logger.Error("failed to unmarshal Lambda event", zap.Error(err))
		return nil, err
	}

	supportedCommands := getSupportedCommands()
	command, ok := supportedCommands[lambdaEvent.Command]
	if !ok {
		return nil, fmt.Errorf("unsupported command: %s. must be 1 of: %v", lambdaEvent.Command, slices.Collect(maps.Keys(supportedCommands)))
	}

	network := lambdaEvent.Network
	supportedNetworks := getSupportedNetworks()
	// All non-Validate commands require caller to specify a target network
	if command != Validate && !slices.Contains(supportedNetworks, network) {
		return nil, fmt.Errorf("invalid network: %s. must be 1 of: %v", network, supportedNetworks)
	}

	// All non-Index commands require caller to specify a timestamp of input file(s) to use
	if command != Index && lambdaEvent.Timestamp == "" {
		return nil, fmt.Errorf("lambda commands require a timestamp of the input file(s) to use")
	}

	// Set TIMESTAMP env var for file I/O prefixes in S3
	var timestamp string
	if command == Index {
		timestamp = time.Now().UTC().Format(time.RFC3339)
	} else {
		timestamp = lambdaEvent.Timestamp
	}
	os.Setenv("TIMESTAMP", timestamp)

	var args []string
	switch command {
	case Index:
		args = []string{"index", "--config", fmt.Sprintf("./local/config-dydx-%s.json", network)}
	case Generate:
		args = []string{"generate", "--config", fmt.Sprintf("./local/config-dydx-%s.json", network)}
	case Override:
		args = []string{"override", "--config", fmt.Sprintf("./local/config-dydx-%s.json", network)}
	case Validate:
		args = []string{"validate", "--market-map", "generated-market-map.json", "--start-delay", "10s", "--duration", "10m", "--enable-all"}
	case Upserts:
		args = []string{"upserts", "--config", fmt.Sprintf("./local/config-dydx-%s.json", network), "--warn-on-invalid-market-map"}
	case Diff:
		args = []string{"diff", "--network", fmt.Sprintf("dydx-%s", network), "--market-map", "generated-market-map.json", "--output", "diff", "--slinky-api"}
	}

	logger.Info("received Lambda command", zap.Strings("args", args))

	return args, nil
}

func lambdaHandler(ctx context.Context, event json.RawMessage) (resp LambdaResponse, err error) {
	logger := logging.Logger(ctx)

	// Fetch CMC API Key from Secrets Manager and set it as env var
	// so it can be used by the Index + Validate jobs
	cmcAPIKey, err := aws.GetSecret(ctx, "market-map-updater-cmc-api-key")
	if err != nil {
		logger.Error("failed to get CMC API key from Secrets Manager", zap.Error(err))
		return resp, err
	}
	os.Setenv("CMC_API_KEY", cmcAPIKey)

	args, err := getArgsFromLambdaEvent(ctx, event)
	if err != nil {
		logger.Error("failed to get args from Lambda event", zap.Error(err))
		return resp, err
	}

	r := createSigningRegistry()
	rootCmd := cmd.RootCmd(r)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		logger.Error("command returned errors", zap.Strings("command", args), zap.Error(err))
		// Return errors for all commands other than "validate".
		// It is expected that "validate" may output errors; these errors do not indicate job failure (ex. transient provider issues, etc.)
		// If these errors are returned from the Lambda handler, the Lambda run will be considered a failure and subsequent jobs in the Step Function will not run.
		if args[0] != "validate" {
			return resp, err
		}
	}

	return LambdaResponse{
		Timestamp: os.Getenv("TIMESTAMP"),
	}, nil
}

func main() {
	if aws.IsLambda() {
		// Running in AWS Lambda
		lambda.Start(lambdaHandler)
	} else {
		// Running locally
		r := createSigningRegistry()
		if err := cmd.RootCmd(r).Execute(); err != nil {
			os.Exit(1)
		}
	}
}
