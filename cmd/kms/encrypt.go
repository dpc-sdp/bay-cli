package kms

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/dpc-sdp/bay-cli/internal/helpers"
	errors "github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func Encrypt(c *cli.Context) error {
	inputContents, err := io.ReadAll(c.App.Reader)
	if err != nil {
		return errors.Wrap(err, "unable to read input")
	}

	if string(inputContents) == "" {
		return errors.New("no input provided")
	}

	// Validate the file is less than 4KB.
	if len(inputContents) > 4096 {
		return errors.New("file exceeds maximum filesize - we plan to support files greater than 4KB in the future")
	}

	keyId, err := helpers.AwsKmsGetKeyIdByTag(map[string]string{
		"project": c.String("project"),
		"key":     c.String("key"),
	})
	if err != nil {
		return err
	}

	client := helpers.AwsKmsClient()
	in := &kms.EncryptInput{
		KeyId:     keyId,
		Plaintext: inputContents,
	}
	out, err := client.Encrypt(context.TODO(), in)
	if err != nil {
		return errors.Wrap(err, "error encrypting payload with key")
	}

	fmt.Fprintf(c.App.Writer, b64.StdEncoding.EncodeToString(out.CiphertextBlob))
	return nil
}
