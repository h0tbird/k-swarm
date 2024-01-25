package cmd

import (
	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// generateInformerCmd
//-----------------------------------------------------------------------------

var generateInformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Outputs informer manifests.",
	Run: func(cmd *cobra.Command, args []string) {

		// Parse the template
		tmpl := parseTemplate("informer")

		// Get the replicas flag
		replicas, _ := cmd.Flags().GetInt("replicas")

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
