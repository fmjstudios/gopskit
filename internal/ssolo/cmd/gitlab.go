package cmd

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/ssolo/app"
	"github.com/spf13/cobra"
)

var _ app.CLIOpt = NewGitLabCommand

func NewGitLabCommand(ssolo *app.State) *cobra.Command {
	var (
		username          string
		password          string
		realm             string
		reflectNamespaces []string
	)

	cmd := &cobra.Command{
		Use:              "gitlab",
		Short:            "Configure GitLab for SAML authentication with Keycloak",
		Long:             "Configure GitLab for SAML authentication with Keycloak as the IdP",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// var exists bool = true // assume true to avoid unneeded creations

			// ctx, cancel := context.WithCancel(context.Background())
			// envF := proc.Must(cmd.Flags().GetString("environment"))
			// environment := proc.Must(core.EnvFromString(envF))
			// namespace := proc.Must(cmd.Flags().GetString("namespace"))
			// label := proc.Must(cmd.Flags().GetString("label"))
			// reflect := proc.Must(cmd.Flags().GetBool("reflect"))

			if password == "" {
				return fmt.Errorf("can't login to Keycloak without password")
			}

			// allow CTRL+C
			// go proc.WaitForCancel(proc.CleanupFunc(func() int {
			// 	cancel()
			// 	return 0
			// }))

			// find pods
			// pods, err := cmdutil.Pods(ssolo, namespace, label)
			// if err != nil {
			// 	return err
			// }

			// // wait until the pod is running
			// kcLeaderPod, err := cmdutil.LeaderPod(ssolo, pods, namespace, label)
			// if err != nil {
			// 	return err
			// }
			// cmdutil.WaitUntilRunning(ssolo, *kcLeaderPod)

			// // port-forward the (leader)
			// ssolo.Log.Infof("Port-forwarding Keycloak leader pod: %s", kcLeaderPod.Name)
			// ctx, cancel := context.WithCancel(context.Background())
			// readyChan := make(chan struct{})
			// stopChan := make(chan struct{})
			// go func(sc, rc chan struct{}) {
			// 	err := ssolo.Kube.PortForward(ctx, *kcLeaderPod, sc, rc)
			// 	if err != nil {
			// 		panic(err)
			// 	}
			// }(stopChan, readyChan)

			// ssolo.Log.Info("Waiting for port-forwarded Keycloak API to become available...")
			// <-readyChan

			// Keycloak Client
			// token, err := ssolo.Keycloak.LoginAdmin(ctx, username, password, realm)
			// if err != nil {
			// 	return err
			// }

			// // persist credentials
			// creds := &cmdutil.Credentials{
			// 	Token: token,
			// 	User: &cmdutil.Login{
			// 		Username: username,
			// 		Password: password,
			// 	},
			// }
			// cmdutil.WriteCredentials(ssolo, environment, creds)

			// _, err = ssolo.Keycloak.CreateRealm(ctx, token.AccessToken, gocloak.RealmRepresentation{
			// 	Realm:   gocloak.StringP("operations"),
			// 	Enabled: gocloak.BoolP(true),
			// })
			// if err != nil {
			// 	return err
			// }

			// secret := &corev1.Secret{
			// 	TypeMeta: metav1.TypeMeta{
			// 		APIVersion: "v1",
			// 		Kind:       "Secret",
			// 	},
			// 	ObjectMeta: metav1.ObjectMeta{
			// 		Name:      "dhparam",
			// 		Namespace: namespace,
			// 		Annotations: map[string]string{
			// 			"reflector.v1.k8s.emberstack.com/reflection-allowed":            "true",
			// 			"reflector.v1.k8s.emberstack.com/reflection-allowed-namespaces": strings.Join(reflectNamespaces, ","),
			// 			"reflector.v1.k8s.emberstack.com/reflection-auto-enabled":       "true",
			// 			"reflector.v1.k8s.emberstack.com/reflection-auto-namespaces":    strings.Join(reflectNamespaces, ","),
			// 		},
			// 	},
			// 	Type: "Opaque",
			// 	Data: map[string][]byte{
			// 		"dhparam.pem": []byte("wtf"),
			// 	},
			// }

			// // check for existence before generating since it takes ages
			// _, err = ssolo.Kube.Secret(secret.Namespace, secret.Name, metav1.GetOptions{})
			// if err != nil {
			// 	exists = false
			// }

			// if exists {
			// 	ssolo.Log.Infof("Skipping creation of Ingress-Nginx Diffie-Hellman parameter secret in namespace: %s. Secret exists: %s", secret.Namespace, secret.Name)
			// 	return nil
			// }

			// ssolo.Kube.CreateSecret(secret.Namespace, secret, metav1.CreateOptions{})
			// ssolo.Log.Infof("Successfully created Ingress-Nginx Diffie-Hellman parameter secret: %s in namespace: %s", secret.Name, secret.Namespace)
			// defer cancel()
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&username, "username", "u", "admin", "The username for the management account within Keycloak")
	cmd.PersistentFlags().StringVarP(&password, "password", "p", "", "The password for the management account within Keycloak")
	cmd.PersistentFlags().StringVarP(&realm, "realm", "r", "master", "The realm for to log into within Keycloak")
	cmd.PersistentFlags().StringArrayVar(&reflectNamespaces, "reflect-namespaces", []string{
		"kube-system",
		"ingress-nginx"},
		"Namespaces to enable for reflection")

	return cmd
}
