package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/util"
)

//-----------------------------------------------------------------------------
// generateInformerCmd
//-----------------------------------------------------------------------------

var generateInformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Outputs informer manifests.",
	Run: func(cmd *cobra.Command, args []string) {

		// Get all the flags
		replicas, _ := cmd.Flags().GetInt("replicas")

		// Parse the template
		tmpl, err := util.ParseTemplate(Assets, "informer")
		if err != nil {
			panic(err)
		}

		// Render the template
		tmpl.Execute(cmd.OutOrStdout(), struct {
			Replicas     int
			NodeSelector string
		}{
			Replicas:     replicas,
			NodeSelector: nodeSelector,
		})
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the informerCmd
	generateCmd.AddCommand(generateInformerCmd)

	// Define the flags
	generateInformerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
	generateInformerCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
	if err := generateInformerCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletionFunc); err != nil {
		panic(err)
	}
}

//-----------------------------------------------------------------------------
// nodeSelectorCompletionFunc
//-----------------------------------------------------------------------------

func nodeSelectorCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"{key1:value1,key2:value2}"}, cobra.ShellCompDirectiveNoFileComp
}
