package cmd

import (

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
)

//-------------------------------------------------------------------------
// installWorkerCmd
//-------------------------------------------------------------------------

var installWorkerCmd = &cobra.Command{
	Use:   "worker [start:end]",
	Short: "Installs worker manifests.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Get the regex
		regex, err := cmd.Flags().GetString("context")
		if err != nil {
			panic(err)
		}

		// Get the contexts that match the regex
		contexts, err := util.GetKubeContexts(regex)
		if err != nil {
			panic(err)
		}

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
