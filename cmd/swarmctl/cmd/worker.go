package cmd

import (

	// Stdlib
	"fmt"

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

		// Read the content from the embedded file system
		content, err := Assets.ReadFile("assets/worker.yaml")
		if err != nil {
			fmt.Println("Error reading the embedded YAML file:", err)
			return
		}

		// Convert the content to a string and print it
		fmt.Println(string(content))
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
