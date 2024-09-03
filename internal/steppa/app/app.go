package app

import (
	"github.com/fmjstudios/gopskit/pkg/cmd"
	"github.com/fmjstudios/gopskit/pkg/logger"
	"github.com/fmjstudios/gopskit/pkg/platform"
	"github.com/fmjstudios/gopskit/pkg/stamp"
	"github.com/spf13/cobra"
)

const (
	Name = "steppa"
)

type Opt func(a *App)

// type App is the implementation for the `steppa` command-line application
type App struct {
	// Name is the name of the application this object is instanced for
	Name string

	// Executor represents a command-line executor to quickly new start processes
	// and closely evaluate their output as well as potential errors
	Executor *cmd.Executor

	// Platform is a utility object performing OS-specific lookups and other
	// data gathering tasks
	Platform *platform.Platform

	// Logger is a wrapper object around Uber's `zap` logger
	Logger *logger.Logger

	// RootCmd is the root command for the cobra command-set. This will be
	// executed to start the command-line application
	RootCmd *cobra.Command

	// Stamps are build-time information that is linked into the final executable
	// by LD. Our Bazel builds stamps builds via LD using the 'x_defs' attribute
	// on the 'go_library' rule
	Stamps *stamp.Stamps
}

// New creates a newly initialized instance of the App type
func New(opts ...Opt) *App {
	platform := platform.New(platform.WithApp(Name))
	logger := logger.New()
	defer logger.Sync()

	exec := cmd.NewExecutor(cmd.WithInheritedEnv())
	stamps := stamp.New()

	a := &App{
		Name:     Name,
		Executor: exec,
		Logger:   logger,
		Platform: platform,
		Stamps:   stamps,
	}

	// (re-)configure
	for _, o := range opts {
		o(a)
	}

	return a
}
