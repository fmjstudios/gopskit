// package tools implements the API to other tools like Helmfile, or Smallstep's 'step' CLI
// Initially I wanted to compile in the direct sources from their respective Go modules
//
// However, Bazel struggles with compilation of the depedencies, mainly due to issues with the
// new 'bzlmod' dependency system
package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Tool represents an executable the gopskit package 'tools' depends on
type Tool string

// String implements the Stringer interface for Tool
func (t *Tool) String() string {
	return string(*t)
}

const (
	gitHubURL = "https://github.com/"
)

var (
	// represents the kubectl CLI
	Kubectl Tool = "kubectl"

	// the helmfile CLI
	Helmfile Tool = "helmfile"

	// the Helm CLI
	Helm Tool = "helm"

	// ref: https://github.com/smallstep/cli
	StepCA Tool = "step"

	// the kustomize CLI
	Kustomize Tool = "kustomize"

	tools = []Tool{Kubectl, Helmfile, Helm, StepCA, Kustomize}
)

// Find checks the system for the required tools. It returns the first error that occurs during the search.
// If any error is returned the map is nil.
func Find() (map[Tool]string, error) {
	t := make(map[Tool]string)

	for _, v := range tools {
		path, err := LookPath(v.String())
		if err != nil {
			return nil, fmt.Errorf("could not find tool: %s. error: %v", v, err)
		}

		t[v] = path
	}

	return t, nil
}

// LookPath is a safer alternative to the direct use of exec.LookPath, which mitigates issues caused by
// non-nil errors due to the fact that the Go runtime resolved the executable to the current directory instead
// of a generic location in $PATH
func LookPath(path string) (string, error) {
	path, err := exec.LookPath(path)
	if errors.Is(err, exec.ErrDot) {
		return path, nil
	}

	return path, err
}

// releaseResponse represents a GitHub API response
type releaseResponse struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
}

// latestGitHubRelease fetches the latest release data for a given GitHub Repository
func latestGitHubRelease(client *resty.Client, url string) (*releaseResponse, error) {
	req := client.R().
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28")

	owner, repo, err := parseGitHubURL(url)
	if err != nil {
		return nil, err
	}

	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := req.Get(api)
	if err != nil {
		return nil, err
	}

	d := &releaseResponse{}
	json.Unmarshal(resp.Body(), d)
	return d, nil
}

// parseGitHubURL parses the repository owner and name from the GitHub URL
// If a non-GitHub URL is given, the function will return an error.
func parseGitHubURL(url string) (string, string, error) {
	var s string

	if strings.Contains(url, gitHubURL) {
		s = strings.TrimPrefix(url, gitHubURL)
	} else {
		return "", "", fmt.Errorf("cannot parse GitHub URL for non-GitHub-URL: %s", url)
	}

	ss := strings.Split(s, "/")
	return ss[0], ss[1], nil
}
