package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	cmdutil "github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/fmjstudios/gopskit/pkg/core"
	"github.com/fmjstudios/gopskit/pkg/proc"
	"github.com/fmjstudios/gopskit/pkg/tools"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

var _ app.CLIOpt = NewInitCommand // assure type compatibility

// NewInitCommand creates the Option which injects the 'init' subcommand into
// the 'waltr' CLI application
func NewInitCommand() func(app *app.State) *cobra.Command {
	return func(app *app.State) *cobra.Command {
		var (
			highAvailability bool
			threshold        int
			shares           int
			secretFile       string
		)

		cmd := &cobra.Command{
			Use:              "initialize",
			Short:            "Initialize Vault",
			Aliases:          []string{"init"},
			Long:             "Initialize Vault runs in High Availability mode",
			TraverseChildren: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				envF := proc.Must(cmd.Flags().GetString("environment"))
				label := proc.Must(cmd.Flags().GetString("label"))
				namespace := proc.Must(cmd.Flags().GetString("namespace"))
				environment := proc.Must(core.EnvFromString(envF))

				var needsUnseal, hasCustomConfig bool
				var vaultNamespace, customConfigName string
				var vaultLeaderPod *corev1.Pod
				var creds *cmdutil.Credentials

				// we unseal by default
				needsUnseal = true

				// look across the entire cluster
				pods, err := cmdutil.Pods(app, "", label)
				if err != nil {
					return fmt.Errorf("could not retrieve Vault pods for label: %s. Error: %v", label, err)
				}

				// ensure we're using a single namespace, or it's passed in
				vaultNamespace, err = cmdutil.EnsureNamespace(pods)
				if err != nil && namespace == "" {
					return fmt.Errorf("found multiple possible Vault pods. the namespace option is unset and %v", err)
				}

				for _, p := range pods {
					app.Log.Debugf("discovered Vault Pod: %s in namespace %s", p.Name, p.Namespace)
					for _, vol := range p.Spec.Volumes {
						if vol.Name == "config" {
							hasCustomConfig = true
							customConfigName = vol.ConfigMap.LocalObjectReference.Name
							break
						}
					}

					if err := cmdutil.DisableAshHistory(app, p); err != nil {
						app.Log.Errorf("could not disable Ash Shell history for Pod: %s. Error: %v", p.Name, err)
					}

					app.Log.Infof("successfully disabled Ash Shell history for Pod: %s", p.Name)
				}

				// The official chart mounts the ConfigMap as 'config' (if the Helm values are set)
				// ref: https://github.com/hashicorp/vault-helm/blob/main/templates/_helpers.tpl#L187
				if hasCustomConfig {
					cm, err := app.Kube.ConfigMap(vaultNamespace, customConfigName, metav1.GetOptions{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
					})

					if err != nil {
						return fmt.Errorf("vault has custom 'config' volume but cannot find corresponding ConfigMap."+
							"Error: %v", err)
					}

					// auto-unseal cannot be configured without this key
					val, ok := cm.Data["extraconfig-from-values.hcl"]
					if ok {
						for _, v := range cmdutil.KnownUnsealTypes {
							if strings.Contains(val, v) {
								needsUnseal = false
								highAvailability = true // must be true if auto-unseal is set
							}
						}

						if !needsUnseal {
							app.Log.Info("found custom chart configuration enabling Auto-Unseal! skipping" +
								" unseal steps")
						}
					}
				} else {
					app.Log.Info("found no configuration ConfigMaps for Vault - proceeding with default steps")
				}

				// wait until the pod is running
				vaultLeaderPod, err = cmdutil.LeaderPod(app, pods, vaultNamespace, label)
				if err != nil {
					return err
				}
				cmdutil.WaitUntilRunning(app, *vaultLeaderPod)

				// port-forward the (leader)
				app.Log.Infof("Port-forwarding Vault Leader: %s", vaultLeaderPod.Name)
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					err := app.Kube.PortForward(ctx, pods[0])
					if err != nil {
						panic(err)
					}
				}()

				// get current status
				status, err := app.VaultClient.System.SealStatus(context.Background())
				if err != nil {
					app.Log.Errorf("could not get Vault status: %v", err)
				}

				// check for initialization
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

					initRes, err := app.VaultClient.System.Initialize(context.Background(), req)
					if err != nil {
						defer cancel()
						return fmt.Errorf("could not initialize Vault instance: %v", err)
					}

					app.Log.Infof("successfully initialized Vault instance: %s", pods[0].Name)

					var data initResponse
					jsn, err := json.MarshalIndent(initRes, "", "")
					if err != nil {
						app.Log.Errorf("could not marshal init request: %v. Vault returned invalid response.", err)
					}
					if err := json.Unmarshal(jsn, &data); err != nil {
						app.Log.Errorf("could not unmarshal init response: %v. Vault returned invalid response.",
							err)
					}

					// determine Vault credentials
					creds = &cmdutil.Credentials{
						Keys:       data.Data.Keys,
						KeysBase64: data.Data.KeysBase64,
						Token:      data.Data.RootToken,
					}

					// write Token somewhere where we can retrieve it
					if secretFile != "" {
						_, err := tools.AddSecretValue(secretFile, map[string]interface{}{
							"vault": map[string]interface{}{
								"token": data.Data.RootToken,
							},
						}, false)

						if err != nil {
							app.Log.Errorf("could not add secret value to file: %s. Error: %v", secretFile, err)
						}
					} else {
						app.Log.Info("secret-file unset. only writing Vault Token to cache path!")
					}

					// always write backup json file to cache path
					err = cmdutil.WriteCredentials(app, environment, creds)
					if err != nil {
						defer cancel()
						return err
					}
				} else {
					app.Log.Info("skipping Vault initialization")
				}
				app.Log.Info("Shutting down Port-forward for Vault Leader Pod")
				cancel()

				// re-read credentials if we skipped initialization
				if status.Data.Initialized {
					creds, err = cmdutil.ReadCredentials(app, environment)
					if err != nil {
						return fmt.Errorf("could not read Vault credentials: %v. Did you initialize Vault with 'waltr'",
							err)
					}
				}

				if err := app.VaultClient.SetToken(creds.Token); err != nil {
					return fmt.Errorf("could not set Vault token: %v", err)
				}

				// unseal the pod(s) - if auto-unseal is not enabled or if we're not initialized yet
				if needsUnseal || !status.Data.Initialized {
					for _, p := range pods {
						err := func() error {
							// wait for boot-up
							for {
								if p.Status.Phase != corev1.PodRunning {
									app.Log.Infof("Vault Pod: %s is not running yet - waiting for Pod to start", p.Name)
									time.Sleep(2500 * time.Millisecond)
								} else {
									app.Log.Infof("Vault Pod: %s is running", p.Name)
									break
								}
							}

							ctx, loopCancel := context.WithCancel(context.Background())
							app.Log.Infof("starting Vault port-forward for pod: %s", p.Name)
							go func() {
								err := app.Kube.PortForward(ctx, p)
								if err != nil {
									panic(err)
								}
							}()

							app.Log.Info("waiting for API's to boot...")
							time.Sleep(time.Millisecond * 2000)

							// unseal
							for i := 0; i < threshold; i++ {
								sealed, err := app.VaultClient.System.SealStatus(context.Background())
								if err != nil {
									loopCancel()
									return fmt.Errorf("could not get Vault status: %v", err)
								}

								if !sealed.Data.Sealed {
									app.Log.Infof("skipping Vault unseal iteration: %d - Vault is unsealed", i)
									continue
								}

								_, err = app.VaultClient.System.Unseal(context.Background(), schema.UnsealRequest{
									Key:   creds.Keys[i],
									Reset: false,
								})

								if err != nil {
									loopCancel()
									return fmt.Errorf("could not unseal Vault instance: %s", p.Name)
								}

								app.Log.Infof("successfully unsealed Vault instance: %s with key %d of threshold %d",
									p.Name, i+1, threshold)
								time.Sleep(time.Millisecond * 250)
							}

							loopCancel()
							return nil
						}()

						if err != nil {
							return err
						}
					}
				} else {
					app.Log.Info("Skipping Vault unseal - Auto-Unseal is enabled.")
				}

				app.Log.Info("successfully initialized Vault")
				return nil
			},
		}

		cmd.PersistentFlags().BoolVar(&highAvailability, "high-availability", false, "Ensure Vault is running in HA mode")
		cmd.PersistentFlags().IntVar(&threshold, "threshold", 4, "The threshold of recovery keys required to unlock Vault")
		cmd.PersistentFlags().IntVar(&shares, "shares", 7, "The amount of total recovery key shares Vault emits")
		cmd.PersistentFlags().StringVar(&secretFile, "secret-file", "",
			"A Helm secrets plugin-encrypted file to inject the token into")

		return cmd
	}
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
