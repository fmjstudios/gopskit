package cmd

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
)

func NewMethodsCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "methods",
		Short:            "auth",
		Aliases:          []string{"auth-methods", "authentication-methods"},
		Long:             "Manage Vault's authentication methods",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
