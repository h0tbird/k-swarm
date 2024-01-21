package cmd

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// informerCmd represents the informer command
//-----------------------------------------------------------------------------

var informerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Generates a swarm informer install manifest and outputs to the console.",
	Run: func(cmd *cobra.Command, args []string) {

		// Read the content from the embedded file system
		content, err := Assets.ReadFile("assets/informer.yaml")
		if err != nil {
			fmt.Println("Error reading the embedded YAML file:", err)
			return
		}

		// Convert the content to a string and print it
		fmt.Println(string(content))
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the informerCmd
	generateCmd.AddCommand(informerCmd)

	// Define the flags
	informerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
}
