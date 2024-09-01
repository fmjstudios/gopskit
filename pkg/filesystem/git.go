package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// RevParseGitRoot fuck u
func RevParseGitRoot(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", nil
	}

	for {
		gitDir := filepath.Join(path, ".git")
		info, err := os.Stat(gitDir)
		if err != nil {
			return "", err
		}

		if !info.IsDir() {
			return "", fmt.Errorf("%s exists but points to a file, rather than a directory", path)
		}

		ok, err := findGitMarkers(gitDir)
		if err != nil {
			return "", err
		}

		if info.IsDir() && ok {
			return filepath.Dir(gitDir), nil
		}

		// check if were in a bare repository
		ok, err = findGitMarkers(path)
		if err != nil {
			return "", err
		}

		if ok {
			return path, nil
		}

		parentDir := filepath.Dir(path)
		if parentDir == path {
			return "", fmt.Errorf("cannot find .git in or below path: %s", path)
		}

		// reset path before new loop iter
		path = parentDir
	}
}

// findGitMarkers checks a given path for the existence of significant directories and files
// related to Git VCS
func findGitMarkers(path string) (bool, error) {
	markers := []string{"config", "HEAD", "branches", "objects", "refs"}
	for _, v := range markers {
		_, err := os.Stat(filepath.Join(path, v))

		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return false, err
			} else {
				return false, nil
			}
		}

		continue
	}

	return true, nil
}
