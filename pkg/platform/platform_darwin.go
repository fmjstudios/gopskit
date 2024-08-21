//go:build darwin

package platform

import "path/filepath"

// determineLogDir uses the previously discovered home directory to determine the directory for
// log files
func (p *Platform) determineLogDir() (string, error) {
	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, "Library", "Application Support", p.app, "Logs")
	return path, nil
}
