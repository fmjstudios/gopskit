// package tools implements the API to other tools like Helmfile, or Smallstep's 'step' CLI
// Initially I wanted to compile in the direct sources from their respective Go modules
//
// However, Bazel struggles with compilation of the depedencies, mainly due to issues with the
// new 'bzlmod' dependency system
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-resty/resty/v2"
)

// Executables represents and executables that either a Tool or another required programs depends on
type Executable int

const (
	kubectl Executable = iota
	helm
	helmfile
	stepCA
	kustomize
)

// String implements the fmt.Stringer interface for the new Executable type
func (e Executable) String() string {
	return [...]string{"kubectl", "helm", "helmfile", "step-ca", "kustomize"}[e]
}

// Index makes the current Executable index retrievable
func (e Executable) Index() int {
	return int(e)
}

var (
	githubURL   = "https://github.com/"
	executables = []Executable{kubectl, helm, helmfile, stepCA, kustomize}
)

// Find checks the system for the required executables. It returns the first error that occurs during
// the search, thereby setting the map return value to nil.
func Find() (map[Executable]string, error) {
	t := make(map[Executable]string)

	for _, v := range executables {
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

// execResult is the result of executing a command via Exec
type execResult struct {
	Command string
	Bin     string
	Args    []string
	StdOut  string
	StdErr  string
	RC      int
	Error   error
}

// Exec executes a command within the current environment (akin to shell) and either returns an
// error if any occurred of an execResult object containing the exit code a copy of the error,
// as well as the string output of the command for StdOut and StdErr
//
// If the program in the command isn't installed we fail fast instead of trying to execute a command
// that does not exist on the system.
//
// Exec allows for the cancellation of a running command via CTRL+C (SIGINT).
func Exec(command ...string) (*execResult, error) {
	var args []string
	var stdOut, stdErr bytes.Buffer
	var cmdErr error = nil

	args = command
	if len(command) <= 1 {
		args = strings.Fields(command[0])
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// allow CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		cancelCtx()
		fmt.Printf("execution of command: '%s' was cancelled via CTRL+C", shortenedCommand(args...))
	}()

	// sanity
	bin, err := LookPath(args[0])
	if err != nil {
		return nil, fmt.Errorf("no such binary or executable in PATH: %s. cannot execute non-existent program", bin)
	}

	cmd := exec.CommandContext(ctx, bin, args[1:]...)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stdErr)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdOut)
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			cmdErr = exitErr
		} else {
			cmdErr = err
		}
	}

	return &execResult{
		Command: shortenedCommand(args...),
		Bin:     bin,
		Args:    args[1:],
		StdOut:  stdOut.String(),
		StdErr:  stdErr.String(),
		RC:      cmd.ProcessState.ExitCode(),
		Error:   cmdErr,
	}, nil
}

// releaseResponse represents a GitHub API response
type releaseResponse struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
}

// latestGitHubRelease fetches the latest release data for a given GitHub Repository. If an authenticated request
// should be made, the passed in client should have a Bearer Auth Token set for every request. To do so, call the
// client.SetAuthToken method before passing it in
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

	if strings.Contains(url, githubURL) {
		s = strings.TrimPrefix(url, githubURL)
	} else {
		return "", "", fmt.Errorf("cannot parse GitHub URL for non-GitHub-URL: %s", url)
	}

	ss := strings.Split(s, "/")
	return ss[0], ss[1], nil
}

// shortenedCommand creates a short representation of a command in order to produce valuable
// debug or info messages without exposing all of the arguments passed in. The func also handles
// cases where entire strings are passed in by splitting the argument if it's only a single one
func shortenedCommand(s ...string) string {
	var strs []string
	if len(s) == 1 {
		strs = strings.Fields(s[0])
	}

	if len(strs) <= 2 {
		return fmt.Sprintf("%s ...", strs)
	} else {
		return fmt.Sprintf("%s ...", strs[:3])
	}
}

// Trim the v from the GitHub Tag
func versionFromTag(tag string) string {
	return strings.TrimPrefix("v", tag)
}

// Add a 'v' to a SemVer version if needed
func tagFromVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}

	return fmt.Sprintf("v%s", version)
}
