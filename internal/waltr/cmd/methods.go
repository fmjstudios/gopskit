package cmd

import (
	"context"
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/fmjstudios/gopskit/pkg/env"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func NewMethodsCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "methods",
		Short:            "Manage Vault's authentication methods",
		Aliases:          []string{"auth", "authentication"},
		Long:             "Manage Vault's authentication methods",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			envF, _ := cmd.Flags().GetString("environment")
			label, _ := cmd.Flags().GetString("label")
			environment := env.FromString(envF)

			pods, err := util.Pods(a, "", label)
			if err != nil {
				return fmt.Errorf("could not retrieve Vault pods for label: %s. Error: %v", label, err)
			}

			//for _, p := range pods {
			//	a.Logger.Infof("discovered Vault Pod: %s in namespace %s", p.Name, p.Namespace)
			//}

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

			// get current methods
			m, err := util.AuthMethods(a)
			if err != nil {
				return fmt.Errorf("could not list enabled Vault authentication methods: %v", err)
			}

			// get API server info
			svc, err := a.KubeClient.Services("default", v1.ListOptions{
				TypeMeta: v1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				LabelSelector: "component=apiserver,provider=kubernetes",
			})

			if err != nil {
				return fmt.Errorf("could not find Kubernetes API server service")
			}

			// enable Kubernetes
			const kp = "kubernetes/"
			if !util.Contains(m, kp) {
				_, err := a.VaultClient.System.AuthEnableMethod(context.Background(), strings.TrimSuffix(kp, "/"),
					schema.AuthEnableMethodRequest{
						Config: map[string]interface{}{
							"kubernetes_host": fmt.Sprintf("https://%s:%d", svc[0].Spec.ClusterIP, svc[0].Spec.Ports[0].Port),
						},
						Description: "authenticate with Kubernetes Service Account Tokens",
						Type:        "kubernetes",
					})

				if err != nil {
					return fmt.Errorf("could not enable Vault authentication method Kubernetes. Error: %v", err)
				}

				a.Logger.Info("Enabled Vault authentication method Kubernetes")
			} else {
				a.Logger.Info("Vault authentication method Kubernetes already enabled")
			}

			// enable OIDC
			const op = "oidc/"
			if !util.Contains(m, op) {
				_, err := a.VaultClient.System.AuthEnableMethod(context.Background(), strings.TrimSuffix(op, "/"),
					schema.AuthEnableMethodRequest{
						Type:        "oidc",
						Description: "authenticate with OpenID Connect",
					})

				if err != nil {
					return fmt.Errorf("could not enable Vault authentication method OIDC. Error: %v", err)
				}

				a.Logger.Info("Enabled Vault authentication method OIDC")
			} else {
				a.Logger.Info("Vault authentication method OIDC already enabled")
			}

			// get current secrets engines
			s, err := util.SecretsEngines(a)
			if err != nil {
				return fmt.Errorf("could not list secrets engines: %v", err)
			}

			// enable KV-V2
			const kvp = "kv/"
			if !util.Contains(s, kvp) {
				_, err := a.VaultClient.System.MountsEnableSecretsEngine(context.Background(),
					strings.TrimSuffix(kvp, "/"),
					schema.MountsEnableSecretsEngineRequest{
						Type:        "kv-v2",
						Description: "store secret values in key/value storage",
						Options: map[string]interface{}{
							"version": "2",
						},
					})

				if err != nil {
					return fmt.Errorf("could not enable Vault secrets engine KV-V2. Error: %v", err)
				}

				a.Logger.Info("Enabled Vault secrets engine KV-V2")
			} else {
				a.Logger.Info("Vault secrets engine KV-V2 already enabled")
			}

			a.Logger.Info("successfully enabled Vault authentication methods: Kubernetes, OIDC")
			a.Logger.Info("successfully enabled Vault secrets engines: KV-V2")
			return nil
		},
	}

	return cmd
}
