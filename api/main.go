package mmuapi

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/skip-mev/connect-mmu/lib/aws"
)

var FILENAME = "latest-mmu-tx.json"

//nolint:unused
func lambdaHandler(_ context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	txJSON, err := aws.ReadFromS3("latest-mmu-tx.json")
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to read %s", FILENAME),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(txJSON),
	}, nil
}

//nolint:unused
func main() {
	lambda.Start(lambdaHandler)
}
