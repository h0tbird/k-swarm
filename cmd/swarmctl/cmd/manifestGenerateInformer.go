package cmd

import (

	// Stdlib
	"text/template"

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

		// Parse embeded template using ParseFS
		tmpl := template.Must(template.ParseFS(Assets, "assets/informer.goyaml"))

		// Get the replicas flag
		replicas, _ := cmd.Flags().GetInt("replicas")

		// Convert the content to a string and print it
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
