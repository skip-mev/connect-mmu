package aws

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var (
	s3Client             *s3.Client
	secretsManagerClient *secretsmanager.Client
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config"))
	}
	s3Client = s3.NewFromConfig(cfg)
	secretsManagerClient = secretsmanager.NewFromConfig(cfg)
}

func IsLambda() bool {
	return os.Getenv("LAMBDA_TASK_ROOT") != ""
}

func GetSecret(ctx context.Context, name string) (secretString string, err error) {
	secretVal, err := secretsManagerClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: aws.String(name)})
	if err != nil {
		return secretString, err
	}
	secretString = aws.ToString(secretVal.SecretString)
	return secretString, nil
}

func ReadFromS3(path string) ([]byte, error) {
	bucket, key, err := getS3Path(path)
	if err != nil {
		return nil, err
	}

	result, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s:%s from S3: %w", bucket, key, err)
	}
	defer result.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	return buf.Bytes(), nil
}

func WriteToS3(path string, bz []byte) error {
	bucket, key, err := getS3Path(path)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(bz),
	})
	if err != nil {
		return fmt.Errorf("failed to write object %s:%s to S3: %w", bucket, key, err)
	}

	err = s3.NewObjectExistsWaiter(s3Client).Wait(
		ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}, time.Minute)
	if err != nil {
		return fmt.Errorf("failed attempt to wait for object %s:%s to exist: %w", bucket, key, err)
	}

	return nil
}

func getS3Path(path string) (bucket string, key string, err error) {
	env := os.Getenv("ENVIRONMENT")
	if env != "staging" && env != "mainnet" {
		return bucket, key, fmt.Errorf("failed to fetch valid 'ENVIRONMENT' from env vars: ENVIRONMENT='%s'", env)
	}
	bucket = env + "-market-map-updater"

	timestamp := os.Getenv("TIMESTAMP")
	pathTokens := strings.Split(path, "/")
	filename := pathTokens[len(pathTokens)-1]
	key = timestamp + "-" + filename

	return bucket, key, nil
}
