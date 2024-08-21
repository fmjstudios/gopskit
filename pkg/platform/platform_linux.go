//go:build linux

package platform

import (
	"path/filepath"
)

// determineLogDir uses the previously discovered home directory to determine the directory for
// log files
func (p *Platform) determineLogDir() (string, error) {
	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".local", "share", p.app, "logs")
	return path, nil
}
