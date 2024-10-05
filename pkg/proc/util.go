package proc

import (
	stderrs "errors"
	"os/exec"
)

// Must ensures the return values of a function are not errors therefore negating
// any further checks that have to be made. Most commonly used as a wrapper within
// new object declarations
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// LookPath is a safer alternative to the direct use of exec.LookPath, which
// mitigates issues caused by non-nil errors due to the fact that the Go runtime
// resolved the executable to the current directory instead of a generic location
// in $PATH
func LookPath(path string) (string, error) {
	path, err := exec.LookPath(path)
	if stderrs.Is(err, exec.ErrDot) {
		return path, nil
	}

	return path, err
}
