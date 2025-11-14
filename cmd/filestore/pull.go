package filestore

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3_types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/urfave/cli/v3"

	"github.com/dpc-sdp/bay-cli/internal/helpers"
)

const (
	S3_OBJECT_METADATA_KEY_CHECKSUM = "sha256"
)

func Pull(ctx context.Context, c *cli.Command) error {
	client := helpers.AwsS3Client()
	verbose := c.Bool("verbose")
	ignoreChecksum := c.Bool("ignore-checksum")
	localPath := c.String("local-path")
	remotePath := c.String("remote-path")
	parsedURL, err := url.Parse(remotePath)
	if err != nil {
		return fmt.Errorf("Error parsing remote-path flag: %v\n", err)
	}
	bucket := parsedURL.Host
	key := parsedURL.Path[1:]

	if verbose {
		fmt.Fprintf(c.ErrWriter, "Pulling s3 object %s to local path %s\n", remotePath, localPath)
	}

	file, err := openOrCreateFile(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if !ignoreChecksum {
		remoteHash, algo, err := getS3ObjectChecksum(ctx, client, bucket, key)
		if err == nil {
			localHash, _ := helpers.GenerateFileChecksum(file, algo)
			if remoteHash == localHash {
				fmt.Fprintf(c.ErrWriter, "File %s matches %s - skipping. You can bypass this check with the --ignore-checksum flag.\n", localPath, remotePath)
				return nil
			}
		}
	}

	// Create a downloader with the S3 client
	downloader := manager.NewDownloader(client)

	// Download the file
	_, err = downloader.Download(context.TODO(), file, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	return err
}

func getS3ObjectChecksum(ctx context.Context, client *s3.Client, bucket, key string) (string, s3_types.ChecksumAlgorithm, error) {
	resp, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return "", "", err
	}

	if val, ok := resp.Metadata[S3_OBJECT_METADATA_KEY_CHECKSUM]; ok {
		return val, s3_types.ChecksumAlgorithmSha256, nil
	}

	return "", "", fmt.Errorf("No checksum found for object %s in bucket %s", key, bucket)
}

func openOrCreateFile(path string) (*os.File, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return file, err
		}
	} else if err != nil {
		return nil, err
	}

	// File exists, open it
	return os.OpenFile(path, os.O_RDWR, 0)
}
