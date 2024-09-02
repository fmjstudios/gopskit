package cmd

import (
	"fmt"
	"os"

	"github.com/fmjstudios/gopskit/internal/fillr/app"
	"github.com/fmjstudios/gopskit/pkg/filesystem"
	"github.com/fmjstudios/gopskit/pkg/util"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	// placeholders for flags values
	var (
		output   string
		template string
	)

	cmd := &cobra.Command{
		Use:   app.APP_NAME + " [FILE]",
		Short: fmt.Sprintf("%s CLI", app.APP_NAME),
		Long:  "Manage authentication for Kubernetes applications using Keycloak",
		Args: func(cmd *cobra.Command, args []string) error {
			// ensure 1 argument
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}

			// ensure the given path exists
			path := args[0]
			exists := filesystem.CheckIfExists(path)
			if !exists {
				return fmt.Errorf("cannot read values input from non-existing file: %s", path)
			}

			return nil
		},
		TraverseChildren: true,
		Example: `
# use built-in template
fillr my-values.yaml -o my-filled-values.yaml

# use a custom template
fillr my-values.yaml -t "{{ index .Values \"kubescape-operator\" \"chartValues\" | get \"%s\" \"%v\" }}" -o my-filled-values.yaml
`,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			values := make(map[string]interface{})
			out := make(map[string]interface{})

			// read input Helm values
			fc, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			err = yaml.Unmarshal(fc, &values)
			if err != nil {
				return err
			}

			// update output map
			util.ReplaceRecursive(values, nil, out, template)

			// render YAML
			yaml, err := yaml.Marshal(out)
			if err != nil {
				return err
			}

			// output to file
			if output != "" {
				if err := os.WriteFile(output, yaml, 0644); err != nil {
					return err
				}
				return nil
			}

			fmt.Println(yaml)
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&output, "output", "o", "", "The file path to output fillr's result to")
	cmd.PersistentFlags().StringVarP(&template, "template", "t", util.REPLACE_TEMPLATE, "The template to use for each YAML key")

	return cmd
}