package cmd

import (
	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// installInformerCmd
//-----------------------------------------------------------------------------

var installInformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Installs informer manifests.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("install informer")
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the informerCmd
	installCmd.AddCommand(installInformerCmd)

	// Define the flags
	installInformerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
}
