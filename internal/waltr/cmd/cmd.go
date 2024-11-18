package cmd

import (
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
	"sync"
)

var (
	// Commands is a slice of CLIOpt options for subcommands of the 'waltr' CLI
	Commands = []app.CLIOpt{
		NewInitCommand,
		NewMountsCommand,
		NewConfigureCommand,
		NewPrepareCommand,
		NewTestCommand,
	}

	// PrepareSubcommands is a slice of CLIOpt options for subcommands of the 'prepare' subcommand
	PrepareSubcommands = []app.CLIOpt{
		NewPrepareKeycloakCommand,
	}
)

func NewRootCommand(waltr *app.State) *cobra.Command {
	var (
		environment string
		label       string
		namespace   string
	)

	cmd := &cobra.Command{
		Use:              app.Name,
		Short:            fmt.Sprintf("%s CLI", app.Name),
		Long:             "Manage HashCorp Vault on Kubernetes",
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		SilenceUsage: true,
	}

	// Kubernetes Flags
	//a.KubeClient.Flags.Namespace = util.StrPtr(app.DefaultNamespace)
	//a.KubeClient.Flags.AddFlags(cmd.PersistentFlags())

	cmd.PersistentFlags().StringVarP(&environment, "environment", "e", "dev", "The execution environment to use (dev, stage, prod)")
	cmd.PersistentFlags().StringVarP(&label, "label", "l", app.DefaultLabel, "The Kubernetes label to filter resources by")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "",
		"The Kubernetes namespace to use. None equates to checking the entire cluster.")

	// add subcommands
	var wg sync.WaitGroup
	wg.Add(len(Commands))
	for _, opt := range Commands {
		go func() {
			cmd.AddCommand(opt()(waltr))
			wg.Done()
		}()
	}
	wg.Wait()

	return cmd
}
