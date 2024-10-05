package cmd

import (
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
)

var _ app.CLIOpt = NewPrepareCommand // assure type compatibility

func NewPrepareCommand() func(app *app.State) *cobra.Command {
	return func(app *app.State) *cobra.Command {
		cmd := &cobra.Command{
			Use:              "prepare",
			Short:            "Prepare Vault for various applications",
			Long:             "Prepare Vault for various applications like GitLab, AWX or Keycloak",
			TraverseChildren: true,
		}

		// subcommands
		for _, subc := range PrepareSubcommands {
			cmd.AddCommand(subc()(app))
		}

		return cmd
	}
}
