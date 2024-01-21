package cmd

import (

	// Stdlib
	"text/template"

	// Community
	"github.com/spf13/cobra"
)

//-------------------------------------------------------------------------
// workerCmd represents the worker command
//-------------------------------------------------------------------------

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Generates a swarm worker install manifest and outputs to the console.",
	Run: func(cmd *cobra.Command, args []string) {

		// Parse embeded template using ParseFS
		tmpl := template.Must(template.ParseFS(Assets, "assets/worker.goyaml"))

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

//-------------------------------------------------------------------------
// init
//-------------------------------------------------------------------------

func init() {

	// Add the command to the workerCmd
	generateCmd.AddCommand(workerCmd)

	// Define the flags
	workerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
}
