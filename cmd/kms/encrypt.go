package kms

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/dpc-sdp/bay-cli/internal/helpers"
	"github.com/urfave/cli/v2"
)

func Encrpyt(c *cli.Context) error {
	keyId := helpers.AwsKmsGetKeyIdByTag(map[string]string{
		"project": c.String("project"),
		"key":     c.String("key"),
	})

	client := helpers.AwsKmsClient()
	in := &kms.EncryptInput{
		KeyId:     keyId,
		Plaintext: []byte("this is nice"),
	}
	out, err := client.Encrypt(context.TODO(), in)
	if err != nil {
		panic(err)
	}
	fmt.Println(b64.StdEncoding.EncodeToString(out.CiphertextBlob))

	return nil
}
