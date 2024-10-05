package cmd

import (
	"fmt"
	"github.com/fmjstudios/gopskit/internal/ssolo/app"
	"github.com/spf13/cobra"
	"sync"
)

var (
	// Commands is a slice of CLIOpt options for subcommands of the 'ssolo' CLI
	Commands = []app.CLIOpt{}
)

func NewRootCommand(ssolo *app.State) *cobra.Command {
	var (
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
			return nil
		},
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	// Kubernetes Flags
	//a.KubeClient.Flags.Namespace = util.StrPtr(app.DefaultNamespace)
	//a.KubeClient.Flags.AddFlags(cmd.PersistentFlags())

	cmd.PersistentFlags().StringVarP(&environment, "environment", "e", "dev", "The execution environment to use (dev, stage, prod)")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "",
		"The Kubernetes namespace to use. None equates to checking the entire cluster.")

	// add subcommands
	var wg sync.WaitGroup
	wg.Add(len(Commands))
	for _, opt := range Commands {
		go func() {
			cmd.AddCommand(opt()(ssolo))
			wg.Done()
		}()
	}
	wg.Wait()

	return cmd
}
