package cmd

import (

	// Stdlib

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

		// Get all the flags
		replicas, _ := cmd.Flags().GetInt("replicas")

		// Parse the template
		tmpl, err := util.ParseTemplate(Assets, "informer")
		if err != nil {
			panic(err)
		}

		// Loop through all clientsets
		for _, clientset := range clientsets {

			// Render the template
			docs, err := util.RenderTemplate(tmpl, struct {
				Replicas int
			}{
				Replicas: replicas,
			})
			if err != nil {
				panic(err)
			}

			// Loop through all yaml documents
			for _, doc := range docs {
				if err := util.ApplyYaml(clientset, doc); err != nil {
					panic(err)
				}
			}
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
