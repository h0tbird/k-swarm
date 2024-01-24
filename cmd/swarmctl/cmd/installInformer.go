package cmd

import (

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
)

//-----------------------------------------------------------------------------
// installInformerCmd
//-----------------------------------------------------------------------------

var installInformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Installs informer manifests.",
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

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add command to rootCmd and informerCmd
	rootCmd.AddCommand(installInformerCmd)
	installCmd.AddCommand(installInformerCmd)

	// Define the flags
	installInformerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
}
