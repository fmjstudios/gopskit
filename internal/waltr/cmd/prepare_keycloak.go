package cmd

import (
	"context"
	"fmt"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/fmjstudios/gopskit/pkg/core"
	"github.com/fmjstudios/gopskit/pkg/helpers"
	"github.com/fmjstudios/gopskit/pkg/proc"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/spf13/cobra"
)

var _ app.CLIOpt = NewPrepareKeycloakCommand // assure type compatibility

func NewPrepareKeycloakCommand(app *app.State) *cobra.Command {
	var (
		token     string
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:              "keycloak",
		Short:            "Prepare Vault for Keycloak",
		Long:             "Prepare Vault with policies and roles for Keycloak",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			envF := proc.Must(cmd.Flags().GetString("environment"))
			label := proc.Must(cmd.Flags().GetString("label"))
			environment := proc.Must(core.EnvFromString(envF))

			pods, err := util.Pods(app, "", label)
			if err != nil {
				return fmt.Errorf("could not retrieve Vault pods for label: %s. Error: %v", label, err)
			}

			if token == "" {
				creds, err := util.ReadCredentials(app, environment)
				if err != nil {
					msg := fmt.Errorf("token option is unset and could not read credentials: %w", err)
					app.Log.Error(msg)
					return err
				}

				token = creds.Token
			}

			// port-forward the (leader)
			app.Log.Infof("Port-forwarding Vault instance: %s", pods[0].Name)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				err := app.Kube.PortForward(ctx, pods[0])
				if err != nil {
					panic(err)
				}
			}()

			// add token
			err = app.VaultClient.SetToken(token)
			if err != nil {
				return fmt.Errorf("could not set token: %v", err)
			}

			// create Kubernetes role
			const r = "keycloak"
			rola, err := util.KubernetesAuthRoles(app)
			if err != nil {
				return err
			}

			if helpers.SliceContains(rola, r) && overwrite || !helpers.SliceContains(rola, r) {
				_, err := app.VaultClient.Auth.KubernetesWriteAuthRole(context.Background(), r,
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

				app.Log.Infof("configured Vault Kubernetes Auth Role %s for Keycloak", r)
			} else {
				app.Log.Infof("skipped configuration of Vault Kubernetes Auth Role %s for Keycloak", r)
			}

			// create Admin credentials
			const adminPath = "keycloak/config"
			adminPass, err := util.GeneratePasswordFromPolicy(app, "alphanumeric-password")
			if err != nil {
				return err
			}

			var adminExists bool = util.HasKvV2Secret(app, adminPath, "kv/")
			if adminExists && overwrite || !adminExists {
				err := util.WriteKvV2Secret(app, adminPath, "kv/", schema.KvV2WriteRequest{
					Data: map[string]interface{}{
						"username": "mg",
						"password": adminPass,
					},
					Options: nil,
				})

				if err != nil {
					return err
				}

				app.Log.Infof("configured Keycloak Admin credentials at path: %s", adminPath)
			} else {
				app.Log.Infof("skipped configuration of Keycloak Admin credentials at path: %s", adminPath)
			}

			// create PostgreSQL credentials (skip if exists)
			const psqlPath = "keycloak/credentials/postgresql"
			psqlPass, err := util.GeneratePasswordFromPolicy(app, "alphanumeric-password")
			if err != nil {
				return err
			}

			var psqlExists bool = util.HasKvV2Secret(app, psqlPath, "kv/")
			if psqlExists && overwrite || !psqlExists {
				err := util.WriteKvV2Secret(app, psqlPath, "kv/", schema.KvV2WriteRequest{
					Data: map[string]interface{}{
						"username": "keycloak",
						"password": psqlPass,
					},
					Options: nil,
				})

				if err != nil {
					return err
				}

				app.Log.Infof("configured Keycloak PostgreSQL credentials at path: %s", psqlPath)
			} else {
				app.Log.Infof("skipped configuration of Keycloak PostgreSQL credentials at path: %s", psqlPath)
			}

			return nil
		},
	}

	addPrepareFlags(cmd, &overwrite, &token)
	return cmd
}
