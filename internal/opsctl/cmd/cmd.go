package cmd

import (
	"fmt"
	"github.com/fmjstudios/gopskit/pkg/platform"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "opsctl",
	Short: "Manage the FMJ Studios Operations Cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := platform.New(platform.WithApp("opsctl"))
		fmt.Printf(`
Home: %s
Config Directory: %s
Cache Directory: %s
Log Directory: %s
Install Directory: %s
Binary: %s
`, p.Home(), p.ConfigDir(), p.CacheDir(), p.LogDir(), p.InstallDir(), p.Bin())

		return nil
	},
}

// Execute
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
