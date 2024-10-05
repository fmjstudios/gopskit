package cmd

import (
	"fmt"
	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/fmjstudios/gopskit/internal/waltr/util"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTestCommand() func(app *app.State) *cobra.Command {
	return func(app *app.State) *cobra.Command {
		var (
			token string
			file  string
		)

		cmd := &cobra.Command{
			Use:              "test",
			Short:            "Temporary test command",
			Aliases:          []string{"t"},
			Long:             "Temporary test command",
			TraverseChildren: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				//envF, _ := cmd.Flags().GetString("environment")
				label, _ := cmd.Flags().GetString("label")
				//environment := env.FromString(envF)

				pods, err := util.Pods(app, "", label)
				if err != nil {
					return fmt.Errorf("could not retrieve Vault pods for label: %s. Error: %v", label, err)
				}

				for _, p := range pods {
					app.Log.Infof("discovered Vault Pod: %s in namespace %s", p.Name, p.Namespace)
				}

				//if token == "" {
				//	creds, err := util.ReadCredentials(a, environment)
				//	if err != nil {
				//		msg := fmt.Errorf("token option is unset and could not read credentials: %w", err)
				//		a.Logger.Error(msg)
				//		return err
				//	}
				//
				//	token = creds.Token
				//}

				// port-forward the (leader)
				//a.Logger.Infof("Port-forwarding Vault instance: %s", pods[0].Name)
				//ctx, cancel := context.WithCancel(context.Background())
				//defer cancel()
				//go func() {
				//	err := a.KubeClient.PortForward(ctx, pods[0])
				//	if err != nil {
				//		panic(err)
				//	}
				//}()

				// add token
				//err = a.VaultClient.SetToken(token)
				//if err != nil {
				//	return fmt.Errorf("could not set token: %v", err)
				//}
				//
				//_, err = tools.AddSecretValue(file, map[string]interface{}{
				//	"vault": map[string]interface{}{
				//		"token": token,
				//	},
				//}, false)
				//
				//if err != nil {
				//	a.Logger.Errorf("could not add secret value to file: %s. Error: %v", file, err)
				//}

				// get API server info
				svc, err := app.Kube.Service("default", "kubernetes", v1.GetOptions{
					TypeMeta: v1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Service",
					},
				})

				if err != nil {
					return fmt.Errorf("could not find Kubernetes API server service")
				}
				app.Log.Info(fmt.Sprintf("https://%s:%d", svc.Spec.ClusterIP, svc.Spec.Ports[0].Port))

				return nil
			},
		}

		cmd.PersistentFlags().StringVarP(&token, "token", "t", "", "The Vault root token")
		cmd.PersistentFlags().StringVarP(&file, "file", "f", "", "The secret file to test")

		return cmd
	}
}
