//go:build linux

package fs

import (
	"path/filepath"
)

// determineLogDir uses the previously discovered home directory to determine the directory for
// log files
func (p *PlatformPaths) determineLogDir() (string, error) {
	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".local", "share", p.AppName, "logs")
	return path, nil
}

// determineLogDir uses the previously discovered home directory to determine the directory for
// program data
func (p *PlatformPaths) determineDataDir() (string, error) {
	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".local", "share", p.AppName, "data")
	return path, nil
}
