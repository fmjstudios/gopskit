package cmd

import (
	"fmt"

	"github.com/fmjstudios/gopskit/internal/waltr/app"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewHACommand(a *app.App) *cobra.Command {
	var (
		high_availability bool
		threshold         int
		shares            int
	)

	cmd := &cobra.Command{
		Use:              "initialize",
		Short:            "Initialize Vault",
		Aliases:          []string{"init"},
		Long:             "Initialize Vault runs in High Availability mode",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pods, err := a.KubeClient.Pods("", metav1.ListOptions{
				LabelSelector: app.DefaultLabel,
			})

			if err != nil {
				panic(err)
			}

			for _, p := range pods {
				fmt.Println("Found Pod:", p.Name, "- in namespace:", p.Namespace)
			}
			// ns, err := a.KubeClient.Namespaces(metav1.ListOptions{})
			// if err != nil {
			// 	fmt.Printf("could not retrieve Kubernetes namespaces: %v\n", err)
			// }

			// for _, n := range ns {
			// 	pods, err := a.KubeClient.Pods(n.Name, metav1.ListOptions{})
			// 	if err != nil {
			// 		fmt.Printf("could not retrieve Pods in namespace: %s. error: %v\n", n.Name, err)
			// 	}

			// 	for _, v := range pods {
			// 		fmt.Printf("Found pod: %s in namespace %s.\n", v.Name, n.Name)
			// 	}
			// }

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&high_availability, "high-availability", false, "Ensure Vault is running in HA mode")
	cmd.PersistentFlags().IntVar(&threshold, "threshold", 4, "The threshold of recovery keys required to unlock Vault")
	cmd.PersistentFlags().IntVar(&shares, "shares", 7, "The amount of total recovery key shares Vault emits")

	return cmd
}
