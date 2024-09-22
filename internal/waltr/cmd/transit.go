package cmd

import (
	"context"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/fmjstudios/gopskit/pkg/env"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/spf13/cobra"
	"strings"
)

func NewTransitCommand(a *app.App) *cobra.Command {
	var (
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:              "transit",
		Short:            "Configure Vault Transit-Encryption",
		Aliases:          []string{"encryption", "transit-encryption"},
		Long:             "Configure Vault for Transit-Encryption with the Vault-Secrets-Operator",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			envF, _ := cmd.Flags().GetString("environment")
			label, _ := cmd.Flags().GetString("label")
			environment := env.FromString(envF)

			pods, err := util.Pods(a, "", label)
			if err != nil {
				return fmt.Errorf("could not retrieve Vault pods for label: %s. Error: %v", label, err)
			}

			creds, err := util.ReadCredentials(a, environment)
			if err != nil {
				msg := fmt.Errorf("could not read credentials: %w", err)
				a.Logger.Error(msg)
				return err
			}

			// port-forward the (leader)
			a.Logger.Infof("Port-forwarding Vault instance: %s", pods[0].Name)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				err := a.KubeClient.PortForward(ctx, pods[0])
				if err != nil {
					panic(err)
				}
			}()

			// add token
			err = a.VaultClient.SetToken(creds.Token)
			if err != nil {
				return fmt.Errorf("could not set token: %v", err)
			}

			s, err := util.SecretsEngines(a)
			if err != nil {
				return err
			}

			// enable transit-encryption
			const tep = "transit/"
			if !util.Contains(s, tep) || overwrite {
				_, err := a.VaultClient.System.MountsEnableSecretsEngine(context.Background(),
					strings.TrimSuffix(tep, "/"),
					schema.MountsEnableSecretsEngineRequest{
						Type:        "transit",
						Description: "encrypt secrets in transit",
					})

				if err != nil {
					return fmt.Errorf("could not enable Vault transit secrets engine. Error: %v", err)
				}

				a.Logger.Info("Enabled Vault transit secrets engine")
			} else {
				a.Logger.Info("Vault secrets engine transit already enabled")
			}

			// create transit secret key (if required)
			const tkp = "transit/keys/vso-client-cache"
			_, err = a.VaultClient.Read(context.Background(), tkp)
			if err != nil {
				// mitigate empty result
				if strings.Contains(err.Error(), "404 Not Found") {
					err = nil
				} else {
					return err
				}
			}

			_, err = a.VaultClient.Write(context.Background(), tkp, map[string]interface{}{})
			if err != nil {
				return err
			}
			a.Logger.Infof("created Vault transit encryption key: %s", tkp)

			// create ACL policy
			pol, err := util.Policies(a)
			if err != nil {
				return err
			}

			const p = "vso-auth"
			if !util.Contains(pol, p) || overwrite {
				_, err := a.VaultClient.System.PoliciesWriteAclPolicy(context.Background(), p,
					schema.PoliciesWriteAclPolicyRequest{
						Policy: fmt.Sprintf(util.ConfigReleasePolicyTemplate, p),
					})
				if err != nil {
					return err
				}

				a.Logger.Infof("configured Vault ACL policy for Transit-Encryption: %s", p)
			} else {
				a.Logger.Infof("skipped configuration of Vault ACL policy for Transit-Encryption: %s", p)
			}

			// create Kubernetes role
			rola, err := util.KubernetesAuthRoles(a)
			if err != nil {
				return err
			}

			if !util.Contains(rola, p) || overwrite {
				_, err := a.VaultClient.Auth.KubernetesWriteAuthRole(context.Background(), p, schema.KubernetesWriteAuthRoleRequest{
					Audience:                      "vault",
					BoundServiceAccountNames:      []string{"vault-secrets-operator"},
					BoundServiceAccountNamespaces: []string{"vault-secrets-operator"},
					TokenPeriod:                   "120",
					TokenPolicies: []string{
						"keycloak",
						"awx",
						"crowdsec",
						"gitlab",
						"gitlab-runner",
						"harbor",
						"headlamp",
						"homepage",
						"jenkins",
						"kubescape",
						"loki",
						"matomo",
						p,
					},
					TokenTtl: "0",
				})

				if err != nil {
					return err
				}

				a.Logger.Infof("configured Vault Kubernetes Auth Role %s for Transit-Encryption", p)
			} else {
				a.Logger.Infof("skipped configuration of Vault Kubernetes Auth Role %s for Transit-Encryption", p)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing configuration")

	return cmd
}
