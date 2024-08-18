package platform

import (
	"github.com/fmjstudios/gopskit/pkg/logger"
	kubehome "k8s.io/client-go/util/homedir"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

var (
	// a package-local logger
	log *logger.Logger = nil

	// global variables
	HomeDir    string
	Executable string
	InstallDir string
	ConfigDir  string
	LogDir     string
)

// set up logger to use within this (shareable) package
func init() {
	log = logger.New()
}

type Option func(p *Platform)

type Platform interface {
	// Home returns the home directory of the current platform
	Home() string

	// Bin returns the path to the current executable
	Bin() string

	// BinPath returns the path the binary installation directory on the local system
	// On Linux it would be something like '/usr/local/bin'
	BinPath() string

	// InstallPath is the path to the installation directory of the local executable
	// The InstallPath may or may not be synonymous with the BinPath
	InstallPath() string

	// ConfigPath is the path to the configuration directory
	ConfigPath() string

	// CacheDir is the caching directory
	CacheDir() string

	// LogDir is the logging the directory
	LogDir() string
}

type Config struct {
	// (global) directories
	homeDir    string
	installDir string
	configDir  string
	cacheDir   string
	logDir     string

	// general info
	appName string // the name of the application we're building paths for

	// relevant paths
	executable string
	binPath    string
	configPath string
	exists     bool
}

// Current returns a fully initialized platform object, using the platform's default paths.
// A call to Current also update the platform-scoped global variables for use outside the package.
//
// Another way of obtaining a Platform object is by using New with the corresponding options.
func Current() *Config {
	return &Config{}
}

// New creates a fully initialized platform object using the specified Options
func New(...Option) *Config {
	return &Config{}
}

// Name returns the pretty name of the current GOOS
func Name() string {
	return runtime.GOOS
}

// Home returns the current user's home directory, respecting the Kubernetes Project's
// sanity checks beforehand
func Home() string {
	home := kubehome.HomeDir()

	if home != "" {
		return home
	}

	home, err := os.UserHomeDir()
	if err == nil {
		return home
	}

	plt := Name()
	usr, _ := user.Current()
	switch plt {
	default:
		home = filepath.Join("home", usr.Username)
	case "windows":
		usrprof := os.Getenv("USERPROFILE")
		if usrprof != "" {
			home = usrprof
		}

		usrdrv := os.Getenv("HOMEDRIVE")

		if usrdrv != "" {
			home = filepath.Join(usrdrv, "Users", usr.Username)
		} else {
			log.Fatal("Could not determine home directory for your system!")
		}
	}

	return home
}
