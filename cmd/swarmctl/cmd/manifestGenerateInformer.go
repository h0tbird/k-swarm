package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
)

//-----------------------------------------------------------------------------
// generateInformerCmd
//-----------------------------------------------------------------------------

var generateInformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Outputs informer manifests.",
	Run: func(cmd *cobra.Command, args []string) {

		// Get all the flags
		replicas, _ := cmd.Flags().GetInt("replicas")

		// Parse the template
		tmpl, err := util.ParseTemplate(Assets, "informer")
		if err != nil {
			panic(err)
		}

		// Render the template
		tmpl.Execute(cmd.OutOrStdout(), struct {
			Replicas int
		}{
			Replicas: replicas,
		})
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the informerCmd
	generateCmd.AddCommand(generateInformerCmd)

	// Define the flags
	generateInformerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
}
