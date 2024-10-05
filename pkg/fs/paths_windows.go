//go:build windows

package fs

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

// determineLogDir uses the Windows-builtin environment variables to look for known
// correct paths for log files on the filesystem
func (p *PlatformPaths) determineLogDir() (string, error) {
	if localAppData != "" {
		return filepath.Join(localAppData, "Logs", p.AppName), nil
	}

	if appData != "" {
		return filepath.Join(appData, "Local", "Logs", p.AppName), nil
	}

	if userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Local", "Logs", p.AppName), nil
	}

	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "AppData", "Local", "Logs", p.AppName), nil
}

// determineDataDir uses the Windows-builtin environment variables to look for known
// correct paths for program data on the filesystem
func (p *PlatformPaths) determineDataDir() (string, error) {
	if programData != "" {
		return filepath.Join(programData, p.AppName), nil
	}

	if homeDrive != "" {
		return filepath.Join(homeDrive, "ProgramData", p.AppName), nil
	}

	if userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Data", "Data", p.AppName), nil
	}

	home, err := p.determineHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "AppData", "Local", "Data", p.AppName), nil
}
