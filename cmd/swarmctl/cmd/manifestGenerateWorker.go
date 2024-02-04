package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"fmt"
	"strconv"
	"strings"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/util"
)

//-------------------------------------------------------------------------
// manifestGenerateWorkerCmd
//-------------------------------------------------------------------------

var manifestGenerateWorkerCmd = &cobra.Command{
	Use:   "worker <start:end>",
	Short: "Outputs worker manifests.",
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
	RunE: func(cmd *cobra.Command, args []string) error {

		// Split args[0] into start and end
		parts := strings.Split(args[0], ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid range format. Please use the format start:end")
		}

		// Convert start and end to integers
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return fmt.Errorf("invalid range. Both start and end should be integers")
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
			}{
				Replicas:     replicas,
				Namespace:    fmt.Sprintf("service-%d", i),
				NodeSelector: nodeSelector,
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
