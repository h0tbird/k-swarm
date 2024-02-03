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
// manifestGenerateInformerCmd
//-----------------------------------------------------------------------------

var manifestGenerateInformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Outputs informer manifests.",
	Run: func(cmd *cobra.Command, args []string) {

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

	// Add the command to the parent
	manifestGenerateCmd.AddCommand(manifestGenerateInformerCmd)
}
