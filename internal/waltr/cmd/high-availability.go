package cmd

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
)

func NewHACommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "high-availability",
		Short:            "ha",
		Aliases:          []string{"high-avail", "availability"},
		Long:             "Ensure Vault runs in High Availability mode",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
