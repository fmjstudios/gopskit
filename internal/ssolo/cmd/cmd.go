package cmd

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/ssolo/app"
	"github.com/spf13/cobra"
)

var (
	// Commands is a slice of CLIOpt options for subcommands of the 'ssolo' CLI
	Commands = []app.CLIOpt{
		NewGitLabCommand,
	}
)

func NewRootCommand(ssolo *app.State) *cobra.Command {
	var (
		label       string
		environment string
		namespace   string
	)

	cmd := &cobra.Command{
		Use:              ssolo.Name,
		Short:            fmt.Sprintf("%s CLI", ssolo.Name),
		Long:             "Manage authentication for Kubernetes applications using Keycloak",
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Usage()
			}

			return nil
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	// Kubernetes Flags
	//a.KubeClient.Flags.Namespace = util.StrPtr(app.DefaultNamespace)
	//a.KubeClient.Flags.AddFlags(cmd.PersistentFlags())

	cmd.PersistentFlags().StringVarP(&label, "label", "l", app.DefaultLabel, "The Kubernetes label to filter resources by")
	cmd.PersistentFlags().StringVarP(&environment, "environment", "e", "dev", "The execution environment to use (dev, stage, prod)")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "",
		"The Kubernetes namespace to use. None equates to checking the entire cluster.")

	// add subcommands
	for _, opt := range Commands {
		cmd.AddCommand(opt(ssolo))
	}

	return cmd
}
