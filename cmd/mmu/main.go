package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

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
	Command string `json:"command"`
}

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

func getArgsFromLambdaEvent(ctx context.Context, event json.RawMessage, cmcApiKey string) ([]string, error) {
	logger := logging.Logger(ctx)

	var lambdaEvent LambdaEvent
	if err := json.Unmarshal(event, &lambdaEvent); err != nil {
		logger.Error("failed to unmarshal Lambda event", zap.Error(err))
		return nil, err
	}

	args := []string{"command"}

	switch command := lambdaEvent.Command; command {
	case "index":
		args[0] = "index"
	case "generate":
		args[0] = "generate"
	case "validate":
		args = []string{"validate", "--market-map", "generated-market-map.json", "--cmc-api-key", cmcApiKey, "--enable-all"}
	case "override":
		args[0] = "override"
	case "upserts":
		args = []string{"upserts", "--warn-on-invalid-market-map"}
	default:
		return nil, fmt.Errorf("received invalid command from Lambda event: %s", command)
	}

	return args, nil
}

func lambdaHandler(ctx context.Context, event json.RawMessage) error {
	logger := logging.Logger(ctx)

	// Fetch CMC API Key from Secrets Manager and set it as env var
	// so it can be used by the Indexer HTTP client
	cmcApiKey, err := aws.GetSecret(ctx, "market-map-updater-cmc-api-key")
	os.Setenv("CMC_API_KEY", cmcApiKey)

	args, err := getArgsFromLambdaEvent(ctx, event, cmcApiKey)
	if err != nil {
		logger.Error("failed to get args from Lambda event", zap.Error(err))
		return err
	}

	r := createSigningRegistry()
	rootCmd := cmd.RootCmd(r)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		logger.Error("failed to execute command", zap.Strings("command", args), zap.Error(err))
		return err
	}

	return nil
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
