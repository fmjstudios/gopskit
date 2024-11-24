package cmd

import (
	"errors"
	"strings"

	"github.com/fmjstudios/gopskit/internal/plattr/app"
	"github.com/fmjstudios/gopskit/pkg/proc"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ app.CLIOpt = NewAWSCommand

func NewAWSCommand(app *app.State) *cobra.Command {
	var (
		accessKeyId       string
		secretAccessKey   string
		reflectNamespaces []string
	)

	cmd := &cobra.Command{
		Use:              "aws",
		Short:            "Configure AWS credentials",
		Long:             "Configure AWS credentials via Kubernetes Secrets",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var annotations = make(map[string]string)
			var exists bool = true

			namespace := proc.Must(cmd.Flags().GetString("namespace"))
			reflect := proc.Must(cmd.Flags().GetBool("reflect"))

			if accessKeyId == "" || secretAccessKey == "" {
				return errors.New("cannot create AWS Secret without credentials")
			}

			if reflect {
				annotations = map[string]string{
					"reflector.v1.k8s.emberstack.com/reflection-allowed":            "true",
					"reflector.v1.k8s.emberstack.com/reflection-allowed-namespaces": strings.Join(reflectNamespaces, ","),
					"reflector.v1.k8s.emberstack.com/reflection-auto-enabled":       "true",
					"reflector.v1.k8s.emberstack.com/reflection-auto-namespaces":    strings.Join(reflectNamespaces, ","),
				}
			}

			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "aws-credentials",
					Namespace:   namespace,
					Annotations: annotations,
				},
				Type: "Opaque",
				Data: map[string][]byte{
					"access_key_id":     []byte(accessKeyId),
					"secret_access_key": []byte(secretAccessKey),
				},
			}

			_, err := app.Kube.Secret(secret.Namespace, secret.Name, metav1.GetOptions{})
			if err != nil {
				exists = false
			}

			if exists {
				app.Log.Infof("Skipping creation of AWS credentials secret in namespace: %s. Secret exists: %s", secret.Namespace, secret.Name)
				return nil
			}

			app.Kube.CreateSecret(secret.Namespace, secret, metav1.CreateOptions{})
			app.Log.Infof("Successfully created secret: %s in namespace: %s", secret.Name, secret.Namespace)
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&accessKeyId, "access-key-id", "", "AWS Access Key ID")
	cmd.PersistentFlags().StringVar(&secretAccessKey, "secret-access-key", "", "AWS Secret Access Key")
	cmd.PersistentFlags().StringArrayVar(&reflectNamespaces, "reflect-namespaces", []string{
		"kube-system",
		"vault",
		"keycloak",
		"gitlab",
		"harbor",
		"awx",
		"matomo"},
		"Namespaces to enable for reflection")

	return cmd
}
