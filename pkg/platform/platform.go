package platform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	kubehome "k8s.io/client-go/util/homedir"
)

var (
	DefaultApplicationName = "gopskit"
	DefaultConfigType      = "yaml"
	KnownConfigTypes       = []string{"yaml", "json"}
)

var _ Config = (*Platform)(nil) // Verify Platform implements Config

// Opt is a configuration option for the PlatformConfig
type Opt func(p *Platform)

type Config interface {
	// Home returns the home directory of the current platform
	Home() string

	// Bin returns the path to the current executable
	Bin() string

	// InstallDir is the path to the installation directory of the local executable
	// If the executable is being run via a SymLink it will refer to the `basename`
	// of the link pointer
	InstallDir() string

	// ConfigDir is the path to the configuration directory
	ConfigDir() string

	// CacheDir is the caching directory
	CacheDir() string

	// LogDir is the logging the directory
	LogDir() string
}

type Platform struct {
	// general info
	app string // the name of the application we're building paths for

	// (global) directories
	homeDir    string
	installDir string
	configDir  string
	cacheDir   string
	logDir     string

	// relevant paths
	executable string
	configType string
	configPath string
	exists     bool
}

// Current returns a fully initialized platform object, using the platform's default paths.
// A call to Current also update the platform-scoped global variables for use outside the package.
//
// Another way of obtaining a Platform object is by using New with the corresponding options.
func Current() *Platform {
	p := &Platform{}
	p.init()
	return p
}

// New creates a fully initialized platform object using the specified Options
func New(opts ...Opt) *Platform {
	bin, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// build and configure the Platform
	p := &Platform{
		app:        DefaultApplicationName,
		executable: bin,
		configType: DefaultConfigType,
	}

	for _, opt := range opts {
		opt(p)
	}

	// find dirs
	p.init()

	// validate the config path
	err = p.findConfigFile()
	if err != nil {
		fmt.Printf("configuration file: %s does not exist\n", p.configPath)
	}

	// keep in line with Windows style decisions
	if runtime.GOOS == "windows" {
		p.app = cases.Title(language.Und, cases.Compact).String(p.app)
	}

	if err := p.init(); err != nil {
		panic(err)
	}

	return p
}

// WithApp configures the Platform object with a custom application name
// By default 'gopskit' will be used
func WithApp(app string) Opt {
	return func(p *Platform) {
		p.app = app
	}
}

// WithConfigPath configures the Platform object with a custom configuration path
// By default this will be determined via the Platforms configDir and the app
func WithConfigPath(path string) Opt {
	return func(p *Platform) {
		p.configPath = path
	}
}

// WithConfigType configures the Platform object with a custom configuration type
// By default we assume the YAML file type since this package is most likely used
// with Kubernetes and 'kubeconfig' files don't have file extensions
//
// If the file does however have an extension the type will be determined using that
func WithConfigType(filetype string) Opt {
	return func(p *Platform) {
		p.configType = filetype
	}
}

// Home returns the determined HOME directory
func (p *Platform) Home() string {
	return p.homeDir
}

// Bin returns current executable or the symlinked path
func (p *Platform) Bin() string {
	return p.executable
}

// InstallDir returns the determined installation directory or basename of the SymLink pointer
func (p *Platform) InstallDir() string {
	return p.installDir
}

// ConfigDir returns the determined configuration directory
func (p *Platform) ConfigDir() string {
	return p.configDir
}

// CacheDir returns the determined cache directory
func (p *Platform) CacheDir() string {
	return p.cacheDir
}

// LogDir returns the determined logging directory
func (p *Platform) LogDir() string {
	return p.logDir
}

// TODO(FMJdev): make this asynchronous
//
// init initializes the new Platform object by calling the needed methods
// to fill the properties
func (p *Platform) init() error {
	var err error
	if p.homeDir, err = p.determineHomeDir(); err != nil {
		return err
	}

	if p.configDir, err = p.determineConfigDir(); err != nil {
		return err
	}

	if p.cacheDir, err = p.determineCacheDir(); err != nil {
		return err
	}

	if p.logDir, err = p.determineLogDir(); err != nil {
		return err
	}

	if p.executable, err = os.Executable(); err != nil {
		return err
	}

	p.installDir = filepath.Dir(p.executable)

	return nil
}

// determineHome determines the current user's home directory, respecting the Kubernetes Project's
// sanity checks beforehand. Darwin uses the exact the function
func (p *Platform) determineHomeDir() (string, error) {
	home := kubehome.HomeDir()

	// use Kubernetes' value if exists
	if home != "" {
		return home, nil
	}

	// still empty? use Go's value then
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return home, nil
}

// determineConfigDir uses the previously discovered home directory to determine a directory for
// configuration files
func (p *Platform) determineConfigDir() (string, error) {
	conf, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(conf, p.app)
	return path, nil
}

// determineCacheDir uses the previously discovered home directory to determine the directory for
// cache files
func (p *Platform) determineCacheDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(cache, p.app)
	return path, nil
}

// findConfigFile tries to find a configuration file in the ConfigDir
func (p *Platform) findConfigFile() error {
	dir := p.ConfigDir()
	p.configPath = fmt.Sprintf("%s/%s.%s", dir, p.app, p.configType)

	if _, err := os.Stat(p.configPath); errors.Is(err, os.ErrNotExist) {
		// it doesn't need to exist since we can operate off of CLI flags/args
		p.exists = false
		return nil
	} else if err != nil {
		p.exists = false
		return fmt.Errorf("could not process  config file: %s", p.configPath)
	} else {
		p.exists = true
		return nil
	}
}
