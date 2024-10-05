// Package proc implements a generic Shell process executor, which allows us to execute
// arbitrary commands from within the program.
package proc

import (
	"bytes"
	"context"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"os/exec"
	"sync"
)

// Opt is a configuration option for the initialization of new Executor's
type Opt func(e *Executor) error

// ExecuteOpt is a runtime option for the Executor's Execute methods
type ExecuteOpt func(e *Executor)

type Executor struct {
	// inheritEnv determines whether the new process should inherit the environment
	// of the starting process
	inheritEnv bool

	// writers are implementations of the io.WriteCloser interface which the
	// command will output the standard streams to. The first element will be
	// standard output, and the second will be standard error. Standard input is
	// omitted since we're looking to execute programs from within an application.
	//
	// NOTE: This value is reset on every call of the Execute method, requiring
	// re-configuration via the ExecutorOpt array.
	writers []io.Writer

	// outputs is an array of file paths we will write the standard output of the
	// started process to. Regardless of how many files there are, all of them will
	// be filled with the standard OUTPUT of the command.
	//
	// NOTE: This value is reset on every call of the Execute method, requiring
	// re-configuration via the ExecutorOpt array.
	outputs []string

	// lock is a Mutex which ensures that only one goroutine may modify the configuration
	lock sync.Mutex

	// exec.Cmd is the embedded Go-native Command to execute
	*exec.Cmd
}

// NewExecutor creates a new Executor with the provided configuration options.
// The method takes in a set of configuration Options which may configure the
// Executor. These Options allow Execute to do things like write files,
// byte-buffers and e.g. inherit the host's environment configuration.
func NewExecutor(opts ...Opt) (*Executor, error) {
	e := &Executor{
		inheritEnv: false,
		lock:       sync.Mutex{},
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	// (re)-configure
	g := new(errgroup.Group)
	for _, o := range opts {
		g.Go(func() error {
			err := o(e)
			if err != nil {
				return err
			}

			return nil
		})
	}

	return e, nil
}

// DefaultExecutor is the default Executor implementation and the base for
// latter configurations via ExecutorOpt's
func DefaultExecutor() *Executor {
	return &Executor{
		inheritEnv: false,
	}
}

// WithInheritedEnv configures the Execute function to inherit the current OS environment
// settings into the command to be executed
func WithInheritedEnv() Opt {
	return func(e *Executor) error {
		e.lock.Lock()
		defer e.lock.Unlock()

		e.inheritEnv = true
		return nil
	}
}

// WithMultiWriters configures the Execute function to use a MultiWriter during execution to
// simultaneously write to both the system's StdOut/Err and to a provided byte-buffer for
// each of those descriptors
func WithMultiWriters(writers ...bytes.Buffer) ExecuteOpt {
	return func(e *Executor) {
		e.lock.Lock()
		defer e.lock.Unlock()

		e.writers = append(e.writers, io.MultiWriter(os.Stdout, &writers[0]))
		e.writers = append(e.writers, io.MultiWriter(os.Stderr, &writers[1]))
	}
}

// WithWriters configures the Execute function to spawn a GoRoutine which fills
// a slice of byte-buffers with date from the StdOut and StdErr output respectively.
func WithWriters(writers ...bytes.Buffer) ExecuteOpt {
	return func(e *Executor) {
		e.lock.Lock()
		defer e.lock.Unlock()

		var bufStdO, bufStdE bytes.Buffer
		e.writers = append(e.writers, &bufStdO, &bufStdE)
	}
}

// WithOutputs configures the Execute function to spawn a GoRoutine which writes
// the standard output of the command to a file on the filesystem
func WithOutputs(paths ...string) ExecuteOpt {
	return func(e *Executor) {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, v := range paths {
			e.outputs = append(e.outputs, v)
		}
	}
}

// Execute executes a system command on the host operating system's shell, however avoids Shell
// expansions like globs etc. The method takes in a set of arguments and a set of configuration
// Options which may configure the execution of the method.
// These Options allow Execute to do things like write files, byte-buffers and e.g. inherit the
// host's environment configuration.
func (e *Executor) Execute(args []string, opts ...ExecuteOpt) ([]string, error) {
	var wg sync.WaitGroup
	e.lock.Lock()
	defer e.lock.Unlock()

	ctx, ctxCancel := context.WithCancel(context.TODO())
	defer ctxCancel()

	g := new(errgroup.Group)
	//args := strings.Fields(command)
	e.Cmd = exec.CommandContext(ctx, args[0], args[1:]...)

	// sanity
	bin, err := LookPath(args[0])
	if err != nil {
		return nil, NotInPathError{Executable: bin}
	}

	// update values for each call
	if err := e.configure(); err != nil {
		return nil, err
	}

	// (re)-configure
	wg.Add(len(opts))
	for _, o := range opts {
		go func() {
			o(e)
			wg.Done()
		}()
	}
	wg.Wait() // config must be done before we do anything

	// create pipes
	readers := []io.ReadCloser{Must(e.StdoutPipe()), Must(e.StderrPipe())}

	// execute command
	err = e.Start()
	if err != nil {
		return nil, err
	}

	// copy command output to generic io.Writers
	for i, wr := range e.writers {
		g.Go(func() error {
			if _, err := copyBytes(readers[i], wr); err != nil {
				return err
			}

			return nil
		})
	}

	// write output to files (if set)
	for _, out := range e.outputs {
		g.Go(func() error {
			var buf bytes.Buffer
			if _, err := copyBytes(readers[0], &buf); err != nil {
				return err
			}

			if err := os.WriteFile(out, buf.Bytes(), 0644); err != nil {
				return err
			}

			return nil
		})
	}

	var bufs []bytes.Buffer
	g.Go(func() error {
		// wait for command to finish
		err := e.Wait()
		if err != nil {
			return err
		}

		// we only need stdout and stderr
		for i := 0; i <= 1; i++ {
			if _, err := copyBytes(readers[i], &bufs[i]); err != nil {
				return err
			}
		}

		return nil
	})

	err = g.Wait()
	if err != nil {
		return nil, ExecuteError{ExitCode: e.ProcessState.ExitCode(), Err: err}
	}

	return []string{
		bufs[0].String(),
		bufs[1].String(),
	}, nil
}

// configure configures the Executor
func (e *Executor) configure() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	e.Dir = wd

	if e.inheritEnv {
		e.Env = append(e.Env, os.Environ()...)
	}

	e.Stdin = nil
	e.writers = []io.Writer{}
	e.outputs = []string{}

	return nil
}

// copyBytes copies bytes from an io.Reader into an io.Writer and returns the entire
// buffers as a whole as well as an error
func copyBytes(src io.Reader, dst io.Writer) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := dst.Write(d)
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
