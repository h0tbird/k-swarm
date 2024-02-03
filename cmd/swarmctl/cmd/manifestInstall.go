package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import "github.com/spf13/cobra"

//-----------------------------------------------------------------------------
// installCmd
//-----------------------------------------------------------------------------

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Generates a manifest and applies it.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the manifestCmd
	manifestCmd.AddCommand(installCmd)

	// Define the flags
	installCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
	installCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
	if err := installCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletionFunc); err != nil {
		panic(err)
	}
}
