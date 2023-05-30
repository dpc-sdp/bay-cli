package helpers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// Return a KMS client with credentials loaded.
func AwsKmsClient() *kms.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	return kms.NewFromConfig(cfg)
}

// Build a KMS alias for a given project/key combination.
func BuildKmsAlias(project, key string) string {
	return fmt.Sprintf("alias/%s-%s", project, key)
}
