package filesystem

import (
	"errors"
	"os"
)

func CheckIfExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
