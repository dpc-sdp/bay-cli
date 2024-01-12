package kms

import (
	"context"
	b64 "encoding/base64"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/dpc-sdp/bay-cli/internal/helpers"
	errors "github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func Decrypt(c *cli.Context) error {
	inputContents, err := io.ReadAll(c.App.Reader)
	if err != nil {
		return errors.Wrap(err, "unable to read input")
	}

	if string(inputContents) == "" {
		return errors.New("no input provided")
	}

	// @todo figure out how to know if this file requires envelope decryption.
	//if len(inputContents) > 4096 {
	//	return errors.New("file exceeds maximum filesize - we plan to support files greater than 4KB in the future")
	//}

	client := helpers.AwsKmsClient()
	decoded, err := b64.StdEncoding.DecodeString(string(inputContents))
	if err != nil {
		return err
	}
	in := &kms.DecryptInput{
		CiphertextBlob: decoded,
	}
	out, err := client.Decrypt(context.TODO(), in)
	if err != nil {
		return errors.Wrap(err, "error decrypting payload")
	}

	io.WriteString(c.App.Writer, string(out.Plaintext))
	return nil
}
