package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/util"
)

//-----------------------------------------------------------------------------
// manifestGenerateInformerCmd
//-----------------------------------------------------------------------------

var manifestGenerateInformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Outputs informer manifests.",
	Example: `
  # Output the generated informer manifest to stdout
  swarmctl manifest generate informer

  # Same using command aliases
  swarmctl m g i

  # Set informer replicas and node selector
  swarmctl m g i --replicas 3 --node-selector '{key1: value1, key2: value2}'
`,
	Aliases: []string{"i"},
	Args:    cobra.ExactArgs(0),
	PreRunE: validateFlags,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Parse the template
		tmpl, err := util.ParseTemplate(Assets, "informer")
		if err != nil {
			return err
		}

		// Render the template
		tmpl.Execute(cmd.OutOrStdout(), struct {
			Replicas     int
			NodeSelector string
		}{
			Replicas:     replicas,
			NodeSelector: nodeSelector,
		})

		// Return
		return nil
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the parent
	manifestGenerateCmd.AddCommand(manifestGenerateInformerCmd)
}
