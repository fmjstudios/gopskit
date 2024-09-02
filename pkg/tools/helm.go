package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/fmjstudios/gopskit/pkg/cmd"
	"github.com/fmjstudios/gopskit/pkg/filesystem"
	"github.com/fmjstudios/gopskit/pkg/util"
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
	e := cmd.NewExecutor(cmd.WithInheritedEnv())
	if len(plugins) == 0 {
		plugins = helmPlugins
	}

	for _, v := range plugins {
		switch v {
		case diff:
			_, _, err := e.Execute(strings.Join([]string{"helm", diff.String(), "version"}, " "))
			if err != nil {
				return err
			}
		case secrets:
			_, _, err := e.Execute(strings.Join([]string{"helm", secrets.String(), "--version"}, " "))
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
	e := cmd.NewExecutor(cmd.WithInheritedEnv())

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
			stdOut, _, err := e.Execute(strings.Join([]string{"helm", diff.String(), "version"}, " "))
			if err != nil {
				return nil, err
			}
			diffVer = stdOut
		case secrets:
			stdOut, _, err := e.Execute(strings.Join([]string{"helm", secrets.String(), "--version"}, " "))
			if err != nil {
				return nil, err
			}
			secretsVer = stdOut
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
	e := cmd.NewExecutor(cmd.WithInheritedEnv())
	if !semver.IsValid(version) {
		return fmt.Errorf("cannot install Helm Plugin %s at invalid version: %v", p.String(), version)
	}

	installArgs := []string{"helm", "plugin", "install", p.String(), helmPluginsRepoMap[p], "--version", tagFromVersion(version)}
	_, _, err := e.Execute(strings.Join(installArgs, " "))
	if err != nil {
		return err
	}

	return nil
}

// func HelmPluginUninstall uninstalls a Helm Plugin
func HelmPluginUninstall(p HelmPlugin) error {
	e := cmd.NewExecutor(cmd.WithInheritedEnv())
	err := ValidateHelmPlugins(p)
	if err != nil {
		return err
	}

	uninstallArgs := []string{"helm", "plugin", "uninstall", p.String()}
	_, _, err = e.Execute(strings.Join(uninstallArgs, " "))
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

// String implements the fmt.Stringer interface for the FileState type
func (f FileState) String() string {
	return [...]string{"encrypted", "decrypted"}[int(f)]
}

// A SOPS-encrypted file always has a 'sops' keys if it's currently encrypted
type SOPSContent struct {
	SOPS SOPSValues `json:"sops" yaml:"sops"`
}

type SOPSValues struct {
	KMS               []yaml.Node `json:"kms" yaml:"kms"`
	GCP_KMS           []yaml.Node `json:"gcp_kms" yaml:"gcp_kms"`
	AZURE_KV          []yaml.Node `json:"azure_kv" yaml:"azure_kv"`
	HC_VAULT          []yaml.Node `json:"hc_vault" yaml:"hc_vault"`
	AGE               []yaml.Node `json:"age" yaml:"age"`
	LastModified      string      `json:"lastmodified" yaml:"lastmodified"`
	Mac               string      `json:"mac" yaml:"mac"`
	PGP               []yaml.Node `json:"pgp" yaml:"pgp"`
	UnencryptedSuffix string      `json:"unencrypted_suffix" yaml:"unencrypted_suffix"`
	Version           string      `json:"version" yaml:"version"`
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
		return encrypted, nil
	}

	return decrypted, nil
}

// EncryptFile encrypts a file using the Helm Secrets Plugin
func EncryptFile(path string) error {
	e := cmd.NewExecutor(cmd.WithInheritedEnv())
	if ok := filesystem.CheckIfExists(path); !ok {
		return fmt.Errorf("cannot encrypt non-existing file: %s", path)
	}

	args := []string{"helm", secrets.String(), "encrypt", "-i", path}
	_, _, err := e.Execute(strings.Join(args, " "))
	if err != nil {
		return err
	}

	return nil
}

// DecryptFile decrypts a file using the Helm Secrets Plugin
func DecryptFile(path string) error {
	e := cmd.NewExecutor(cmd.WithInheritedEnv())
	if ok := filesystem.CheckIfExists(path); !ok {
		return fmt.Errorf("cannot decrypt non-existing file: %s", path)
	}

	args := []string{"helm", secrets.String(), "encrypt", "-i", path}
	_, _, err := e.Execute(strings.Join(args, " "))
	if err != nil {
		return err
	}

	return nil
}

// GetSecretValue parses the file at the provided path, first checking whether it actually exists.
// If it does we check if it's encrypted and decrypt it if required. Afterwards the YAML file contents
// are read an returned via a JSONPath
func GetSecretValue(path, jsonPath string, unencrypted bool) (string, error) {
	var data yaml.Node

	if ok := filesystem.CheckIfExists(path); !ok {
		return "", fmt.Errorf("cannot get value from non-existing file: %s", path)
	}

	state, err := GetFileState(path)
	fmt.Printf("GetSecretValue file state is: %s\n", state)
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

	if err := yaml.Unmarshal(content, &data); err != nil {
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

	if len(nodes) > 1 {
		return "", fmt.Errorf("JSONPath expression: %s did not result in unique YAML node", jsonPath)
	}

	return nodes[0].Value, nil
}

// AddSecretValue initializes a data map to an existing data object
func AddSecretValue(path string, data map[string]interface{}, unencrypted bool) (map[string]interface{}, error) {
	var root map[string]interface{}
	var comp = map[string]interface{}{
		"secrets": data,
	}

	if ok := filesystem.CheckIfExists(path); !ok {
		return nil, fmt.Errorf("cannot add value to non-existing file: %s", path)
	}

	state, err := GetFileState(path)
	if err != nil {
		return nil, err
	}

	if state == encrypted {
		if err := DecryptFile(path); err != nil {
			return nil, err
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, err
	}

	if err := util.DeepMergeMap(root, comp); err != nil {
		return nil, err
	}

	if !unencrypted {
		if err := EncryptFile(path); err != nil {
			return nil, err
		}
	}

	return root, nil
}
