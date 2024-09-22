package cmd

import (
	"context"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/fmjstudios/gopskit/pkg/env"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/spf13/cobra"
)

func NewConfigureCommand(a *app.App) *cobra.Command {
	var (
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:              "configure",
		Short:            "Configure Vault",
		Aliases:          []string{"conf", "config"},
		Long:             "Configure ACL and password policies within Vault",
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

			// get current policies
			pol, err := util.Policies(a)
			fmt.Println("Got ACL policies:", pol)
			if err != nil {
				return err
			}

			// get current password policies
			ppol, err := util.PasswordPolicies(a)
			fmt.Println("Got password policies:", ppol)
			if err != nil {
				return err
			}

			// current policies for Helm releases
			for _, v := range util.Releases {
				if !util.Contains(pol, v) || overwrite {
					_, err := a.VaultClient.System.PoliciesWriteAclPolicy(context.Background(), v, schema.PoliciesWriteAclPolicyRequest{
						Policy: fmt.Sprintf(util.ConfigReleasePolicyTemplate, v),
					})
					if err != nil {
						return err
					}

					a.Logger.Infof("configured Vault ACL policy for Helm release %s ", v)
				} else {
					a.Logger.Infof("skipped configuration of Vault ACL policy for Helm release %s ", v)
				}
			}

			// configure Vault policies to use for authenticated users
			for k, v := range util.ConfigAclPolicies {
				if !util.Contains(pol, k) {
					_, err := a.VaultClient.System.PoliciesWriteAclPolicy(context.Background(), k,
						schema.PoliciesWriteAclPolicyRequest{
							Policy: v,
						})
					if err != nil {
						return err
					}

					a.Logger.Infof("configured Vault ACL policy %s ", k)
				} else {
					a.Logger.Infof("skipped configuration of Vault ACL policy %s ", k)
				}
			}

			// configure Vault password policies to generate secure secrets later on
			for k, v := range util.ConfigPasswordPolicies {
				if !util.Contains(ppol, k) {
					_, err := a.VaultClient.System.PoliciesWritePasswordPolicy(context.Background(), k,
						schema.PoliciesWritePasswordPolicyRequest{
							Policy: v,
						})
					if err != nil {
						return err
					}

					a.Logger.Infof("configured Vault password policy %s ", k)
				} else {
					a.Logger.Infof("skipped configuration of Vault password policy %s ", k)
				}
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing configuration")

	return cmd
}
