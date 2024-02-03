package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import "github.com/spf13/cobra"

//-----------------------------------------------------------------------------
// manifestGenerateCmd
//-----------------------------------------------------------------------------

var manifestGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a manifest and outputs it.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the parent
	manifestCmd.AddCommand(manifestGenerateCmd)

	// --replicas flag
	manifestGenerateCmd.PersistentFlags().IntVar(&replicas, "replicas", 1, "Number of replicas to deploy.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("replicas", replicasCompletionFunc); err != nil {
		panic(err)
	}

	// --node-selector flag
	manifestGenerateCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletionFunc); err != nil {
		panic(err)
	}
}
