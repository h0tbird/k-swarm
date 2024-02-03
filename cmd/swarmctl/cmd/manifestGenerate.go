package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import "github.com/spf13/cobra"

//-----------------------------------------------------------------------------
// generateCmd
//-----------------------------------------------------------------------------

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a manifest and outputs it.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the manifestCmd
	manifestCmd.AddCommand(generateCmd)

	// Define the flags
	generateCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
	generateCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
	if err := generateCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletionFunc); err != nil {
		panic(err)
	}
}
