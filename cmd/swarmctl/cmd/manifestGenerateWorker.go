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
// generateWorkerCmd
//-------------------------------------------------------------------------

var generateWorkerCmd = &cobra.Command{
	Use:   "worker <start:end>",
	Short: "Outputs worker manifests.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Get all the flags
		replicas, _ := cmd.Flags().GetInt("replicas")

		// Split args[0] into start and end
		parts := strings.Split(args[0], ":")
		if len(parts) != 2 {
			fmt.Println("Invalid range format. Please use the format start:end.")
			return
		}

		// Convert start and end to integers
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			fmt.Println("Invalid range. Both start and end should be integers.")
			return
		}

		// Parse the template
		tmpl, err := util.ParseTemplate(Assets, "worker")
		if err != nil {
			panic(err)
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
	},
}

//-------------------------------------------------------------------------
// init
//-------------------------------------------------------------------------

func init() {

	// Add the command to the workerCmd
	generateCmd.AddCommand(generateWorkerCmd)

	// Define the flags
	generateWorkerCmd.PersistentFlags().Int("replicas", 1, "Number of replicas to deploy.")
	generateWorkerCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
}
