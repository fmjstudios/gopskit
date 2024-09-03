package cmd

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/steppa/app"
	"github.com/spf13/cobra"
)

func NewRootCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              app.Name,
		Short:            fmt.Sprintf("%s CLI", app.Name),
		Long:             "Generate and update Smallstep PKI information",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
