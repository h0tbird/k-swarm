package cmd

import (
	// Community
	"github.com/spf13/cobra"
)

//-------------------------------------------------------------------------
// installWorkerCmd
//-------------------------------------------------------------------------

var installWorkerCmd = &cobra.Command{
	Use:   "worker [start:end]",
	Short: "Installs worker manifests.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("install worker")
	},
}

//-------------------------------------------------------------------------
// init
//-------------------------------------------------------------------------

func init() {

	// Add the command to the installCmd
	installCmd.AddCommand(installWorkerCmd)

	// Define the flags
	installWorkerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
}
