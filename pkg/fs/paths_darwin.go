//go:build darwin

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

	path := filepath.Join(home, "Library", "Application Support", p.AppName, "Logs")
	return path, nil
}

// determineLogDir uses the previously discovered home directory to determine the directory for
// program data
func (p *PlatformPaths) determineDataDir() (string, error) {
	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, "Library", p.AppName, "Data")
	return path, nil
}
