//go:build windows

package platform

import (
	"os"
	"path/filepath"
)

var (
	// General Windows paths
	userProfile = os.Getenv("USERPROFILE")
	homeDrive   = os.Getenv("HOMEDRIVE") // most likely 'C:'

	// Windows application paths
	appData      = os.Getenv("APPDATA")
	localAppData = os.Getenv("LOCALAPPDATA")
	programData  = os.Getenv("PROGRAMDATA")
)

// determineLogDir uses the previously discovered home directory to determine the directory for
// log files
func (p *Platform) determineLogDir() (string, error) {
	if localAppData != "" {
		return filepath.Join(localAppData, "Logs", p.app), nil
	}

	if userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Local", "Logs", p.app), nil
	}

	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "AppData", "Local", "Logs", p.app), nil
}
