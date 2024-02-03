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
	Use:   "informer",
	Short: "Installs informer manifests.",
	RunE: func(cmd *cobra.Command, args []string) error {

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
			}{
				Replicas:     replicas,
				NodeSelector: nodeSelector,
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
