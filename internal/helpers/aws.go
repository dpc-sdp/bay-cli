package helpers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
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
