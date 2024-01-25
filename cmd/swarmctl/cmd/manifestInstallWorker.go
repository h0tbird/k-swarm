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

		// TODO: Do something
		for _, context := range contexts {
			cmd.Println(context)
		}
	},
}

//-------------------------------------------------------------------------
// init
//-------------------------------------------------------------------------

func init() {

	// Add command to rootCmd installCmd
	rootCmd.AddCommand(installWorkerCmd)
	installCmd.AddCommand(installWorkerCmd)

	// Define the flags
	installWorkerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
}
