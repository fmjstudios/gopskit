package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/fmjstudios/gopskit/pkg/env"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/spf13/cobra"
	"time"
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
			envF, _ := cmd.Flags().GetString("environment")
			label, _ := cmd.Flags().GetString("label")
			environment := env.FromString(envF)

			pods, err := util.Pods(a, "", label)
			if err != nil {
				return fmt.Errorf("could not retrieve Vault pods for label: %s. Error: %v", label, err)
			}

			for _, p := range pods {
				a.Logger.Infof("discovered Vault Pod: %s in namespace %s", p.Name, p.Namespace)

				if err := util.DisableAshHistory(a, p); err != nil {
					a.Logger.Errorf("could not disable Ash Shell history for Pod: %s. Error: %v", p.Name, err)
				}

				a.Logger.Infof("successfully disabled Ash Shell history for Pod: %s", p.Name)
			}

			// port-forward the (leader)
			a.Logger.Infof("Port-forwarding Vault instance: %s", pods[0].Name)
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				err := a.KubeClient.PortForward(ctx, pods[0])
				if err != nil {
					panic(err)
				}
			}()

			// get current status
			status, err := a.VaultClient.System.SealStatus(context.Background())
			if err != nil {
				a.Logger.Errorf("could not get Vault status: %v", err)
			}

			// check for initialization
			var creds *util.Credentials
			if !status.Data.Initialized {
				var req schema.InitializeRequest
				req = schema.InitializeRequest{
					SecretShares:    int32(shares),
					SecretThreshold: int32(threshold),
				}

				// HA requires the use of recovery keys, Shamir doesn't support it however
				if highAvailability {
					req = schema.InitializeRequest{
						RecoveryShares:    int32(shares),
						RecoveryThreshold: int32(threshold),
					}
				}

				initRes, err := a.VaultClient.System.Initialize(context.Background(), req)
				if err != nil {
					cancel()
					return fmt.Errorf("could not initialize Vault instance: %v", err)
				}

				a.Logger.Infof("successfully initialized Vault instance: %s", pods[0].Name)

				var data initResponse
				jsn, err := json.MarshalIndent(initRes, "", "")
				if err != nil {
					a.Logger.Errorf("could not marshal init request: %v. Vault returned invalid response.", err)
				}
				if err := json.Unmarshal(jsn, &data); err != nil {
					a.Logger.Errorf("could not unmarshal init response: %v. Vault returned invalid response.", err)
				}

				creds = &util.Credentials{
					Keys:       data.Data.Keys,
					KeysBase64: data.Data.KeysBase64,
					Token:      data.Data.RootToken,
				}
				err = util.WriteCredentials(a, environment, creds)
			} else {
				a.Logger.Debug("skipping Vault initialization")
				creds, err = util.ReadCredentials(a, environment)
				if err != nil {
					cancel()
					return err
				}
			}

			a.Logger.Info("Shutting down main port-forward for Vault")
			cancel()

			// unseal the pod(s)
			for _, p := range pods {
				err := func() error {
					ctx, loopCancel := context.WithCancel(context.Background())
					defer loopCancel()

					a.Logger.Infof("starting Vault port-forward for pod: %s", p.Name)
					go func() {
						err := a.KubeClient.PortForward(ctx, p)
						if err != nil {
							panic(err)
						}
					}()

					a.Logger.Info("waiting for API's to boot...")
					time.Sleep(time.Second * 5)

					// unseal
					for i := 0; i < threshold; i++ {
						sealed, err := a.VaultClient.System.SealStatus(context.Background())
						if err != nil {
							return fmt.Errorf("could not get Vault status: %v", err)
						}

						if !sealed.Data.Sealed {
							a.Logger.Debugf("skipping Vault unseal iteration: %d - Vault is unsealed", i)
							continue
						}

						_, err = a.VaultClient.System.Unseal(context.Background(), schema.UnsealRequest{
							Key:   creds.Keys[i],
							Reset: false,
						})

						if err != nil {
							return fmt.Errorf("could not unseal Vault instance: %s", p.Name)
						}

						a.Logger.Infof("successfully unsealed Vault instance: %s with key %d of threshold %d",
							p.Name, i+1, threshold)
					}

					return nil
				}()

				if err != nil {
					return err
				}
			}

			if err != nil {
				return fmt.Errorf("could not write credentials: %v", err)
			}

			a.Logger.Info("successfully saved Vault credentials")
			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&highAvailability, "high-availability", false, "Ensure Vault is running in HA mode")
	cmd.PersistentFlags().IntVar(&threshold, "threshold", 4, "The threshold of recovery keys required to unlock Vault")
	cmd.PersistentFlags().IntVar(&shares, "shares", 7, "The amount of total recovery key shares Vault emits")

	return cmd
}

// vault initialization output
type initResponse struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	LeaseDuration int    `json:"lease_duration"`
	Renewable     bool   `json:"renewable"`
	Data          struct {
		Keys       []string `json:"keys"`
		KeysBase64 []string `json:"keys_base64"`
		RootToken  string   `json:"root_token"`
	} `json:"data"`
	Warnings any `json:"warnings"`
}
