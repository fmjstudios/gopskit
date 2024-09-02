package common

import (
	"github.com/fmjstudios/gopskit/pkg/cmd"
	"github.com/fmjstudios/gopskit/pkg/kube"
	"github.com/fmjstudios/gopskit/pkg/logger"
	"github.com/fmjstudios/gopskit/pkg/platform"
	"github.com/fmjstudios/gopskit/pkg/stamp"
	"github.com/spf13/cobra"
)

// GOpsKitApp is the common central application type embedded by most, if not
// all application's specific types (within `internal`)
type GOpsKitApp struct {
	// Name is the name of the application this object is instanced for
	Name string

	// Executor represents a command-line executor to quickly new start processes
	// and closely evaluate their output as well as potential errors
	Executor *cmd.Executor

	// KubeClient is a Kubernetes Client capable of performing all Kubernetes
	// operations, initialized from the same sources as `kubectl`
	KubeClient *kube.KubeClient

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
