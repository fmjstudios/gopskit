package proc

import (
	stderrs "errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/fmjstudios/gopskit/pkg/log"
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

type CleanupFunc func() int

func WaitForCancel(cleanup CleanupFunc) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// block until a signal is received
	<-c
	log.Global.Info("CTRL+C received. Cancelling current operation...")
	code := cleanup()
	os.Exit(code)
}
