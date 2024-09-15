package cmd

import (
	"context"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/hashicorp/vault-client-go/schema"
	"time"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
)

func NewHACommand(a *app.App) *cobra.Command {
	var (
		highAvailability bool
		threshold        int
		shares           int
	)

	cmd := &cobra.Command{
		Use:              "initialize",
		Short:            "Initialize Vault",
		Aliases:          []string{"init"},
		Long:             "Initialize Vault runs in High Availability mode",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			label, _ := cmd.Flags().GetString("label")

			pods, err := util.Pods(a, "", label)
			if err != nil {
				panic(err)
			}

			for _, p := range pods {
				fmt.Println("Found Pod:", p.Name, "- in namespace:", p.Namespace)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				err := a.KubeClient.PortForward(ctx, pods[0])
				if err != nil {
					panic(err)
				}
			}()

			status, err := a.VaultClient.System.Initialize(context.Background(), schema.InitializeRequest{
				PgpKeys:           nil,
				RecoveryPgpKeys:   nil,
				RecoveryShares:    0,
				RecoveryThreshold: 0,
				RootTokenPgpKey:   "",
				SecretShares:      0,
				SecretThreshold:   0,
				StoredShares:      0,
			})

			if err != nil {
				panic(err)
			}

			fmt.Println("Vault status:", status.Data)
			fmt.Println("Done")

			fmt.Println("Sleeping for 30 seconds")
			time.Sleep(30 * time.Second)

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&highAvailability, "high-availability", false, "Ensure Vault is running in HA mode")
	cmd.PersistentFlags().IntVar(&threshold, "threshold", 4, "The threshold of recovery keys required to unlock Vault")
	cmd.PersistentFlags().IntVar(&shares, "shares", 7, "The amount of total recovery key shares Vault emits")

	return cmd
}
