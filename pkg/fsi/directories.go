package fs

import (
	"os"
)

// TempDir creates a temporary directory within the default temporary directory
// configured for your OS. On Unix it's likely '/tmp'.
func TempDir(pattern string) (string, error) {
	path, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", err
	}

	return path, nil
}

// TempFile
func TempFile(pattern string) (*os.File, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, err
	}

	return file, nil
}
