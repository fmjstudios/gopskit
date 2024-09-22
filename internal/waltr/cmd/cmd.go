package cmd

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
)

func NewRootCommand(a *app.App) *cobra.Command {
	var (
		environment string
		label       string
	)

	cmd := &cobra.Command{
		Use:              app.Name,
		Short:            fmt.Sprintf("%s CLI", app.Name),
		Long:             "Set up and manage HashiCorp's Vault on Kubernetes",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Kubernetes Flags
	//a.KubeClient.Flags.Namespace = util.StrPtr(app.DefaultNamespace)
	//a.KubeClient.Flags.AddFlags(cmd.PersistentFlags())

	cmd.PersistentFlags().StringVarP(&environment, "environment", "e", "dev", "The execution environment to use (dev, stage, prod)")
	cmd.PersistentFlags().StringVarP(&label, "label", "l", app.DefaultLabel, "The Kubernetes label to filter resources by")

	// subcommands
	cmd.AddCommand(NewHACommand(a))
	cmd.AddCommand(NewMethodsCommand(a))
	cmd.AddCommand(NewConfigureCommand(a))
	cmd.AddCommand(NewTransitCommand(a))
	cmd.AddCommand(NewPrepareCommand(a))

	return cmd
}
