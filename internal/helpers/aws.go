package helpers

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"strings"
)

// Return a KMS client with credentials loaded.
func AwsKmsClient() *kms.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	return kms.NewFromConfig(cfg)
}

// Returns a list of customer-managed KMS keys that are enabeld.
func AwsKmsListEnabledCustomerKeys() []types.KeyListEntry {
	client := AwsKmsClient()
	allKeys, err := client.ListKeys(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	keys := make([]types.KeyListEntry, 0)
	for _, key := range allKeys.Keys {
		in := &kms.DescribeKeyInput{KeyId: key.KeyId}
		keyInfo, err := client.DescribeKey(context.TODO(), in)
		if err != nil {
			panic(err)
		}

		if keyInfo.KeyMetadata.Enabled {
			if keyInfo.KeyMetadata.KeyManager == types.KeyManagerTypeCustomer {
				keys = append(keys, key)
			}
		}
	}
	return keys
}

// Returns a KMS key ID by tags.
func AwsKmsGetKeyIdByTag(selectors map[string]string) (*string, error) {
	client := AwsKmsClient()
	keys := AwsKmsListEnabledCustomerKeys()
	for _, key := range keys {
		in := &kms.ListResourceTagsInput{
			KeyId: key.KeyId,
		}
		tagsOnKey, err := client.ListResourceTags(context.TODO(), in)
		if err != nil {
			return nil, err
		}
		for _, tag := range tagsOnKey.Tags {
			for tagKey, tagValue := range selectors {
				if tagKey == strings.ToLower(*tag.TagKey) && tagValue == strings.ToLower(*tag.TagValue) {
					return key.KeyId, nil
				}
			}
		}
	}
	return nil, errors.New("unable to find requested kms key")
}
