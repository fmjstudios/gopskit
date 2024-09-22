package cmd

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
)

func NewPrepareCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "prepare",
		Short:            "Prepare Vault for various applications",
		Long:             "Prepare Vault for various applications like GitLab, AWX or Keycloak",
		TraverseChildren: true,
	}

	// subcommands
	cmd.AddCommand(NewPrepareKeycloakCommand(a))

	return cmd
}
