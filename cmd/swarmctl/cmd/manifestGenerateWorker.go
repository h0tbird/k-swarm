package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/util"
)

//-------------------------------------------------------------------------
// manifestGenerateWorkerCmd
//-------------------------------------------------------------------------

var manifestGenerateWorkerCmd = &cobra.Command{
	Use:          "worker <start:end>",
	Short:        "Outputs worker manifests.",
	SilenceUsage: true,
	Example: `
	  # Output the generated workers 1 to 1 manifests
	  swarmctl manifest generate worker 1:1

	  # Same using command aliases
	  swarmctl m g w 1:1

	  # Set worker replicas and node selector
	  swarmctl m g w 1:1 --replicas 3 --node-selector '{key1: value1, key2: value2}'
`,
	Aliases: []string{"w"},
	Args:    cobra.ExactArgs(1),
	PreRunE: validateFlags,
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		// Set the error prefix
		cmd.SetErrPrefix("\nError:")

		// Parse the range
		start, end, err = util.ParseRange(args[0])
		if err != nil {
			return err
		}

		// Parse the template
		tmpl, err := util.ParseTemplate(Assets, "worker")
		if err != nil {
			return err
		}

		// Loop from start to end
		for i := start; i <= end; i++ {

			// Render the template
			tmpl.Execute(cmd.OutOrStdout(), struct {
				Replicas     int
				Namespace    string
				NodeSelector string
				Version      string
				ImageTag     string
			}{
				Replicas:     replicas,
				Namespace:    fmt.Sprintf("service-%d", i),
				NodeSelector: nodeSelector,
				Version:      version,
				ImageTag:     imageTag,
			})
		}

		// Return
		return nil
	},
}

//-------------------------------------------------------------------------
// init
//-------------------------------------------------------------------------

func init() {

	// Add the command to the parent
	manifestGenerateCmd.AddCommand(manifestGenerateWorkerCmd)
}
