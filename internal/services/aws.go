package services

import (
	"context"
	"expo-open-ota/config"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var (
	s3Client     *s3.Client
	initS3Client sync.Once
)

func GetS3Client() (*s3.Client, error) {
	var err error

	initS3Client.Do(func() {
		var cfg aws.Config
		opts := []func(*awsconfig.LoadOptions) error{}

		// Für Google Cloud Storage spezielle Konfiguration
		baseEndpoint := config.GetEnv("AWS_BASE_ENDPOINT")
		if baseEndpoint == "https://storage.googleapis.com" {
			// GCS-spezifische Konfiguration
			opts = append(opts, awsconfig.WithRegion("auto"))
		} else {
			// Standard AWS-Konfiguration
			opts = append(opts, awsconfig.WithRegion(config.GetEnv("AWS_REGION")))
		}

		accessKey := config.GetEnv("AWS_ACCESS_KEY_ID")
		secretKey := config.GetEnv("AWS_SECRET_ACCESS_KEY")
		if accessKey != "" && secretKey != "" {
			opts = append(opts, awsconfig.WithCredentialsProvider(
				aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
					return aws.Credentials{
						AccessKeyID:     accessKey,
						SecretAccessKey: secretKey,
					}, nil
				}),
			))
		}

		cfg, err = awsconfig.LoadDefaultConfig(context.TODO(), opts...)
		if err == nil {
			if baseEndpoint != "" {
				s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
					o.BaseEndpoint = aws.String(baseEndpoint)
					// Für Google Cloud Storage: Spezielle Konfiguration
					if baseEndpoint == "https://storage.googleapis.com" {
						o.UsePathStyle = true
					}
				})
			} else {
				s3Client = s3.NewFromConfig(cfg)
			}
		}
	})

	if err != nil {
		return nil, fmt.Errorf("error loading AWS configuration: %w", err)
	}
	return s3Client, nil
}

func FetchSecret(secretName string) string {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	resp, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		log.Fatalf("Failed to retrieve secret %s: %v", secretName, err)
	}

	if resp.SecretString == nil {
		log.Fatalf("Secret %s has no SecretString", secretName)
	}

	return *resp.SecretString
}
