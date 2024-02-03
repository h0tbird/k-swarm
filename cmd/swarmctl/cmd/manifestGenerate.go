package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (
	// Community
	"github.com/spf13/cobra"
)

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

//-----------------------------------------------------------------------------
// nodeSelectorCompletionFunc
//-----------------------------------------------------------------------------

func nodeSelectorCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"{key1:value1,key2:value2}"}, cobra.ShellCompDirectiveNoFileComp
}
