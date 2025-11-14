package helpers

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	s3_types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"hash"
	"hash/crc32"
	"io"
	"os"
)

// GenerateFileChecksum computes a file checksum using the specified algorithm
// Supported algorithms are SHA256, SHA1, CRC32, and CRC32C
// Returns the checksum as a hex string and any error encountered
func GenerateFileChecksum(file *os.File, algorithm s3_types.ChecksumAlgorithm) (string, error) {
	defer file.Seek(0, 0)
	var h hash.Hash

	switch algorithm {
	case s3_types.ChecksumAlgorithmSha256:
		h = sha256.New()
	case s3_types.ChecksumAlgorithmSha1:
		h = sha1.New()
	case s3_types.ChecksumAlgorithmCrc32:
		h = crc32.NewIEEE()
	case s3_types.ChecksumAlgorithmCrc32c:
		h = crc32.New(crc32.MakeTable(crc32.Castagnoli))
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// For CRC32 and CRC32C, we convert the uint32 to hex
	if algorithm == s3_types.ChecksumAlgorithmCrc32 || algorithm == s3_types.ChecksumAlgorithmCrc32c {
		checksum := h.Sum(nil)
		return hex.EncodeToString(checksum), nil
	}

	// For other hash algorithms
	checksum := h.Sum(nil)
	return hex.EncodeToString(checksum), nil
}
