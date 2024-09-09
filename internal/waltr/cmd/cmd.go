package cmd

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/pkg/util"
	"github.com/spf13/cobra"
)

func NewRootCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              app.Name,
		Short:            fmt.Sprintf("%s CLI", app.Name),
		Long:             "Set up and manage HashiCorp's Vault on Kubernetes",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(NewHACommand(a))
	cmd.AddCommand(NewMethodsCommand(a))
	a.KubeClient.Flags.Namespace = util.StrPtr(app.DefaultNamespace)
	a.KubeClient.Flags.AddFlags(cmd.PersistentFlags())
	return cmd
}
