package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/util"
)

//-----------------------------------------------------------------------------
// manifestInstallInformerCmd
//-----------------------------------------------------------------------------

var manifestInstallInformerCmd = &cobra.Command{
	Use:          "informer",
	Short:        "Installs informer manifests.",
	SilenceUsage: true,
	Example: `
  # Install the informer to the current context
  swarmctl manifest install informer

  # Same using command aliases
  swarmctl m i i

  # Same using a shoret command chain
  swarmctl informer

  # Same using a short command chain with aliases
  swarmctl i

  # Install the informer to a specific context
  swarmctl i --context my-context

  # Install the informer to all contexts that match a regex
  swarmctl i --context 'my-.*'

  # Install the informer to all contexts that match a regex and set the replicas
  swarmctl i --context 'my-.*' --replicas 3

  # Install the informer to all contexts that match a regex and set the node selector
  swarmctl i --context 'my-.*' --node-selector '{key1: value1, key2: value2}'
`,
	Aliases: []string{"i"},
	Args:    cobra.ExactArgs(0),
	PreRunE: validateFlags,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Set the error prefix
		cmd.SetErrPrefix("\nError:")

		// Parse the template
		tmpl, err := util.ParseTemplate(Assets, "informer")
		if err != nil {
			return err
		}

		// Loop through all contexts
		for name, context := range contexts {

			// Print the context
			fmt.Printf("\n%s\n\n", name)

			// Render the template
			docs, err := util.RenderTemplate(tmpl, struct {
				Replicas     int
				NodeSelector string
				Version      string
			}{
				Replicas:     replicas,
				NodeSelector: nodeSelector,
				Version:      version,
			})
			if err != nil {
				return err
			}

			// Loop through all yaml documents
			for _, doc := range docs {
				if err := context.ApplyYaml(doc); err != nil {
					return err
				}
			}
		}

		// Return
		return nil
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add command to rootCmd and informerCmd
	rootCmd.AddCommand(manifestInstallInformerCmd)
	manifestInstallCmd.AddCommand(manifestInstallInformerCmd)
}
