package helpers

import "os"

// Validates a file is smaller than the given maxSize parameter.
func FileValidateSize(filepath string, maxSize int) (bool, error) {
	fi, err := os.Stat(filepath)
	if err != nil {
		return false, err
	}
	size := fi.Size()
	return int(size) < maxSize, nil
}
