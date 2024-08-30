package tools

import (
	"fmt"
	"os"

	"github.com/fmjstudios/gopskit/pkg/filesystem"
	"github.com/go-resty/resty/v2"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

// HelmPlugin represents a Helm plugin required for gopskit to work
type HelmPlugin int

const (
	diff HelmPlugin = iota
	secrets
)

// String implements the fmt.Stringer interface for the new HelmPlugin type
func (p HelmPlugin) String() string {
	return [...]string{"diff", "secrets"}[p]
}

// Index makes the current HelmPlugin index
func (p HelmPlugin) Index() int {
	return int(p)
}

var (
	helmPlugins        = []HelmPlugin{diff, secrets}
	helmPluginsRepoMap = map[HelmPlugin]string{
		diff:    "https://github.com/databus23/helm-diff",
		secrets: "https://github.com/jkroepke/helm-secrets",
	}
)

// ValidateHelmPlugins checks if the required Helm Plugins "diff" and "secrets" are currently installed
func ValidateHelmPlugins(plugins ...HelmPlugin) error {
	// use (built-in) const if no args are passed
	if len(plugins) == 0 {
		plugins = helmPlugins
	}

	for _, v := range plugins {
		switch v {
		case diff:
			_, err := Exec("helm", diff.String(), "version")
			if err != nil {
				return err
			}
		case secrets:
			_, err := Exec("helm", secrets.String(), "--version")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// HelmPluginVersion retrieves the versions for all or some of the required Helm Plugins
func HelmPluginVersion(plugins ...HelmPlugin) (map[HelmPlugin]string, error) {
	var diffVer, secretsVer string

	// use (built-in) const if no args are passed
	if len(plugins) == 0 {
		plugins = helmPlugins
	}

	// sanity
	if err := ValidateHelmPlugins(plugins...); err != nil {
		return nil, err
	}

	// we can safely ignore errors if we get here
	for _, v := range plugins {
		switch v {
		case diff:
			res, _ := Exec("helm", diff.String(), "version")
			diffVer = res.StdOut
		case secrets:
			res, _ := Exec("helm", secrets.String(), "--version")
			secretsVer = res.StdOut
		}
	}

	return map[HelmPlugin]string{
		diff:    diffVer,
		secrets: secretsVer,
	}, nil
}

// HelmPluginRequiresUpdate determines if an update to a plugin is required
func HelmPluginRequiresUpdate(token string, plugins ...HelmPlugin) (map[HelmPlugin]bool, error) {
	req := make(map[HelmPlugin]bool)

	httpC := resty.New()
	if token != "" {
		httpC.SetAuthToken(token)
	}

	// use (built-in) const if no args are passed
	if len(plugins) == 0 {
		plugins = helmPlugins
	}

	// sanity
	if err := ValidateHelmPlugins(plugins...); err != nil {
		return nil, err
	}

	for _, v := range plugins {
		pluginVersion, _ := HelmPluginVersion(v)
		remoteTag, err := latestGitHubRelease(httpC, helmPluginsRepoMap[v])
		if err != nil {
			return nil, err
		}

		remoteVer := versionFromTag(remoteTag.TagName)
		req[v] = semver.Compare(pluginVersion[v], remoteVer) == -1
	}

	return req, nil
}

// func HelmPluginInstall installs a Helm Plugin from it's remote source
func HelmPluginInstall(p HelmPlugin, version string) error {
	if !semver.IsValid(version) {
		return fmt.Errorf("cannot install Helm Plugin %s at invalid version: %v", p.String(), version)
	}

	installArgs := []string{"helm", "plugin", "install", p.String(), helmPluginsRepoMap[p], "--version", tagFromVersion(version)}
	_, err := Exec(installArgs...)
	if err != nil {
		return err
	}

	return nil
}

// func HelmPluginUninstall uninstalls a Helm Plugin
func HelmPluginUninstall(p HelmPlugin) error {
	err := ValidateHelmPlugins(p)
	if err != nil {
		return err
	}

	uninstallArgs := []string{"helm", "plugin", "uninstall", p.String()}
	_, err = Exec(uninstallArgs...)
	if err != nil {
		return err
	}

	return nil
}

type FileState int

const (
	encrypted FileState = iota
	decrypted
)

// // String implements the fmt.Stringer interface for the new FileState type
func (f FileState) String() string {
	return [...]string{"encrypted", "decrypted"}[int(f)]
}

// A SOPS-encrypted file always has a 'sops' keys if it's currently encrypted
type SOPSContent struct {
	SOPS yaml.Node `json:"sops" yaml:"sops"`
}

// GetFileState checks the contents of a file for existing SOPS encryption and returns the
// current state of the file
func GetFileState(path string) (FileState, error) {
	var c SOPSContent

	if ok := filesystem.CheckIfExists(path); !ok {
		return -1, fmt.Errorf("cannot get state of non-existing file: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return -1, err
	}

	if err := yaml.Unmarshal(content, c); err != nil {
		return decrypted, nil
	}

	return encrypted, nil
}

// EncryptFile encrypts a file using the Helm Secrets Plugin
func EncryptFile(path string) error {
	if ok := filesystem.CheckIfExists(path); !ok {
		return fmt.Errorf("cannot encrypt non-existing file: %s", path)
	}

	args := []string{"helm", "secrets", "encrypt", "-i", path}
	_, err := Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

// DecryptFile decrypts a file using the Helm Secrets Plugin
func DecryptFile(path string) error {
	if ok := filesystem.CheckIfExists(path); !ok {
		return fmt.Errorf("cannot decrypt non-existing file: %s", path)
	}

	args := []string{"helm", "secrets", "encrypt", "-i", path}
	_, err := Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

// GetSecretValue
func GetSecretValue(path, jsonPath string, unencrypted bool) (string, error) {
	var data yaml.Node

	if ok := filesystem.CheckIfExists(path); !ok {
		return "", fmt.Errorf("cannot geg value from non-existing file: %s", path)
	}

	state, err := GetFileState(path)
	if err != nil {
		return "", err
	}

	if state == encrypted {
		if err := DecryptFile(path); err != nil {
			return "", err
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(content, data); err != nil {
		return "", err
	}

	yp, err := yamlpath.NewPath(jsonPath)
	if err != nil {
		return "", err
	}

	nodes, err := yp.Find(&data)
	if err != nil {
		return "", err
	}

	if !unencrypted {
		if err := EncryptFile(path); err != nil {
			return "", err
		}
	}

	return nodes[0].Value, nil
}