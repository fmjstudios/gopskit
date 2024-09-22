package cmd

import (
	"context"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/fmjstudios/gopskit/pkg/env"
	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/spf13/cobra"
	"strings"
)

func NewPrepareKeycloakCommand(a *app.App) *cobra.Command {
	var (
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:              "keycloak",
		Short:            "Prepare Vault for Keycloak",
		Long:             "Prepare Vault with policies and roles for Keycloak",
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

			// create Kubernetes role
			const r = "keycloak"
			rola, err := util.KubernetesAuthRoles(a)
			if err != nil {
				return err
			}

			if !util.Contains(rola, r) || overwrite {
				_, err := a.VaultClient.Auth.KubernetesWriteAuthRole(context.Background(), r,
					schema.KubernetesWriteAuthRoleRequest{
						Audience:                      "vault",
						BoundServiceAccountNames:      []string{r},
						BoundServiceAccountNamespaces: []string{r},
						TokenPeriod:                   "24h",
						TokenPolicies: []string{
							r,
						},
						TokenTtl: "0",
					})

				if err != nil {
					return err
				}

				a.Logger.Infof("configured Vault Kubernetes Auth Role %s for Keycloak", r)
			} else {
				a.Logger.Infof("skipped configuration of Vault Kubernetes Auth Role %s for Keycloak", r)
			}

			// create PostgreSQL credentials
			const psqlPath = "keycloak/credentials/postgresql"
			psqlPass, err := util.GeneratePasswordFromPolicy(a, "alphanumeric-password")
			if err != nil {
				return err
			}

			var exists bool
			_, err = a.VaultClient.Secrets.KvV2Read(context.Background(), psqlPath, vault.WithMountPath("kv/"))
			if err != nil {
				// mitigate empty result
				if strings.Contains(err.Error(), "404 Not Found") {
					err = nil
					exists = false
				}
			} else {
				exists = true
			}

			if !exists {
				// write
				_, err = a.VaultClient.Secrets.KvV2Write(context.Background(), psqlPath, schema.KvV2WriteRequest{
					Data: map[string]interface{}{
						"username": "keycloak",
						"password": psqlPass,
					},
					Options: nil,
				}, vault.WithMountPath("kv/"))

				if err != nil {
					return err
				}

				a.Logger.Infof("configured Keycloak PostgreSQL credentials at path: %s", psqlPath)
			} else {
				a.Logger.Infof("skipped configuration of Keycloak PostgreSQL credentials at path: %s", psqlPath)
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing configuration")

	return cmd
}
