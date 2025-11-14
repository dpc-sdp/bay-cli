package helpers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	AwsDefaultRegion string = "ap-southeast-2"
)

// Return a KMS client with credentials loaded.
func AwsKmsClient() *kms.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if cfg.Region == "" {
		cfg.Region = AwsDefaultRegion
	}
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	return kms.NewFromConfig(cfg)
}

// Build a KMS alias for a given project/key combination.
func BuildKmsAlias(project, key string) string {
	return fmt.Sprintf("alias/%s-%s", project, key)
}

// Return a S3 client with credentials loaded.
func AwsS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if cfg.Region == "" {
		cfg.Region = AwsDefaultRegion
	}
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	return s3.NewFromConfig(cfg)
}
