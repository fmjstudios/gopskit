package cmd

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
)

func NewRootCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              app.APP_NAME,
		Short:            fmt.Sprintf("%s CLI", app.APP_NAME),
		Long:             "Set up and manage HashiCorp's Vault on Kubernetes",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(NewHACommand(a))
	cmd.AddCommand(NewMethodsCommand(a))

	return cmd
}
