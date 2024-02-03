package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (
	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// manifestCmd
//-----------------------------------------------------------------------------

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "The manifest command generates swarm manifests.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(manifestCmd)
}

//-----------------------------------------------------------------------------
// nodeSelectorCompletionFunc
//-----------------------------------------------------------------------------

func nodeSelectorCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"{key1:value1,key2:value2}"}, cobra.ShellCompDirectiveNoFileComp
}
