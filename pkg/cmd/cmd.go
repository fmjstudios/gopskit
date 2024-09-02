package cmd

import (
	"bytes"
	"context"
	stderrs "errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fmjstudios/gopskit/pkg/errors"
	"github.com/fmjstudios/gopskit/pkg/filesystem"
)

type ExecutorOpt func(e *Executor) error
type ExecutorExecuteOpt func(e *Executor)

type Executor struct {
	// InheritEnv determines whether the new process should inhert the environment
	// of the starting process
	InheritEnv bool

	// Chan is a general channel used to notify GoRoutines
	Chan chan struct{}

	// StopChan is a channel notifying receivers of potential errors that occurred
	// during the completion of the process
	StopChan chan error

	// Writers are implementations of the io.WriteCloser interface which the
	// command will output the standard streams to. The first element will be
	// standard output, and the second will be standard error. Standard input is
	// omitted since we're looking to execute programs from within an application.
	//
	// NOTE: This value is reset on every call of the Execute method, requiring
	// re-configuration via the ExecutorOpt array.
	Writers []io.Writer

	// OutputFiles is an array of files we will write the standard output of the
	// started process to. Regardless of how many files there are, all of them will
	// be filled with the standard OUTPUT of the command.
	//
	// NOTE: This value is reset on every call of the Execute method, requiring
	// re-configuration via the ExecutorOpt array.
	OutputFiles []*os.File

	// exec.Cmd is the embedded Go-native Command to execute
	*exec.Cmd
}

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

var (
	// DefaultExecutor is the default Executor implementation and the base for
	// latter configurations via ExecutorOpt's
	DefaultExecutor = &Executor{
		InheritEnv: false,
		Chan:       make(chan struct{}, 1),
		StopChan:   make(chan error, 1),
	}
)

// WithInheritedEnv configures the Execute function to inherit the current OS environment
// settings into the command to be exeucted
func WithInheritedEnv() ExecutorOpt {
	return func(e *Executor) error {
		e.InheritEnv = true
		return nil
	}
}

// WithMultiWriters configures the Execute function to use a MultiWriter during execution to
// simultaneously write to both the system's StdOut/Err and to a provided byte-buffer for
// each of those descriptors
func WithMultiWriters(writers ...bytes.Buffer) ExecutorExecuteOpt {
	return func(e *Executor) {
		e.Writers = append(e.Writers, io.MultiWriter(os.Stdout, &writers[0]))
		e.Writers = append(e.Writers, io.MultiWriter(os.Stderr, &writers[1]))
	}
}

// WithWriters configures the Execute function to spawn a GoRoutine which fills
// a slice of byte-buffers with date from the StdOut and StdErr output respectively.
func WithWriters(writers ...bytes.Buffer) ExecutorExecuteOpt {
	return func(e *Executor) {
		var bufStdO, bufStdE bytes.Buffer
		e.Writers = append(e.Writers, &bufStdO, &bufStdE)
	}
}

// WithOutputFiles configures the Execute function to spawn a GoRoutine which which copies
// the StdOut output of the command to a file on the filesystem
func WithOutputFiles(paths ...string) ExecutorExecuteOpt {
	return func(e *Executor) {
		for _, v := range paths {
			f := Must(filesystem.CreateFile(v))
			e.OutputFiles = append(e.OutputFiles, f)
		}
	}
}

// NewExecutor creates a new Executor with the provided configuration options.
// The method takes in a set of configuration Options which may configure the
// Executor. These Options allow Execute to do things like write files,
// byte-buffers and e.g. inherit the host's environment configuration.
func NewExecutor(opts ...ExecutorOpt) *Executor {
	e := &Executor{
		InheritEnv: false,
		Chan:       make(chan struct{}, 1),
		StopChan:   make(chan error, 2),
	}
	// configure
	for _, o := range opts {
		if err := o(e); err != nil {
			panic(err)
		}
	}
	return e
}

// Execute executes a system command on the host operating system's shell, however avoids Shell
// expansions like globs etc. The method takes in a set of arguments and a set of configuration
// Options which may configure the execution of the method.
// These Options allow Execute to do things like write files, byte-buffers and e.g. inherit the
// host's environment configuration.
func (e *Executor) Execute(command string, opts ...ExecutorExecuteOpt) (string, string, error) {
	ctx, ctxCancel := context.WithCancel(context.TODO())
	defer ctxCancel()

	args := strings.Fields(command)
	e.Cmd = exec.CommandContext(ctx, args[0], args[1:]...)

	// sanity
	bin, err := LookPath(args[0])
	if err != nil {
		return "", "", errors.NotInPATH(bin)
	}

	// update values for each call
	if err := updateCmd(e); err != nil {
		return "", "", err
	}

	// configure
	for _, o := range opts {
		o(e)
	}

	// create pipes
	stdOutR := Must(e.StdoutPipe())
	stdErrR := Must(e.StderrPipe())

	// notify GoRoutines of command start
	err = e.Start()
	e.Chan <- struct{}{}
	if err != nil {
		return "", "", errors.Execute(e.ProcessState.ExitCode(), err)
	}

	// check for writers
	if len(e.Writers) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(e.Writers))

		for i := 0; i < len(e.Writers); i++ {
			go func(startChan chan struct{}, stopChan chan error, readers []io.ReadCloser) {
				defer wg.Done()
				<-startChan

				Must(copyBytes(e.Writers[i], readers[i]))

				stopChan <- nil
			}(e.Chan, e.StopChan, []io.ReadCloser{stdOutR, stdErrR})
		}

		wg.Wait()
	} else {
		e.StopChan <- nil
	}

	// check for output files
	if len(e.OutputFiles) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(e.OutputFiles))

		for i := 0; i < len(e.OutputFiles); i++ {
			go func(startChan chan struct{}, stopChan chan error, stdOut io.ReadCloser, files []*os.File) {
				defer wg.Done()
				<-startChan

				var buf bytes.Buffer
				Must(copyBytes(&buf, stdOut))

				if err := filesystem.WriteFile(files[i], buf.Bytes()); err != nil {
					stopChan <- err
				}

				stopChan <- nil
			}(e.Chan, e.StopChan, stdOutR, e.OutputFiles)
		}

		wg.Wait()
	} else {
		e.StopChan <- nil
	}

	// wait for command to finish
	if cmdErr := e.Wait(); cmdErr != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			err = exitErr
		} else {
			err = cmdErr
		}
	}

	// wait for goroutines
	execErr := <-e.StopChan
	if execErr != nil {
		err = execErr
	}

	if err != nil {
		return "", "", errors.Execute(e.ProcessState.ExitCode(), err)
	}

	ctxCancel()
	return "", "", nil
}

// updateCmd ...
func updateCmd(e *Executor) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	e.Dir = wd

	if e.InheritEnv {
		e.Env = append(e.Env, os.Environ()...)
	}

	e.Stdin = nil
	e.Writers = []io.Writer{}
	e.OutputFiles = []*os.File{}

	return nil
}

// copyBytes ...
func copyBytes(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
