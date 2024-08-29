package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// func CreateFile(path string, data []byte) error {

// }

// func CreateDirectory(path string, data []byte) error {

// }

// Remove removes all files and directories at the specified path. Before running delete operations the
// function checks if the path actually exists, if it does not, it exits silently. If an error occurs
// during the delete operation, that isn't a fs.PathError, we panic.
func Remove(path string) error {
	// don't delete non-existing paths
	exists := CheckIfExists(path)
	if !exists {
		return nil
	}

	err := os.RemoveAll(path)
	isPathErr := errors.Is(err, &fs.PathError{})

	if err != nil {
		if !isPathErr {
			panic(fmt.Errorf("os.RemoveAll returned unknown error: %v", err))
		}

		return err
	}

	return nil
}

// CheckIfExists checks if the specified path exists and returns the boolean result
func CheckIfExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
