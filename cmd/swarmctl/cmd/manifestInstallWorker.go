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
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
)

//-------------------------------------------------------------------------
// installWorkerCmd
//-------------------------------------------------------------------------

var installWorkerCmd = &cobra.Command{
	Use:   "worker <start:end>",
	Short: "Installs worker manifests.",
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

		// Loop through all configs
		for _, config := range configs {

			// Get the clients
			client, err := util.GetClient(config)
			if err != nil {
				panic(err)
			}

			// Loop trough all services
			for i := start; i <= end; i++ {

				// Render the template
				docs, err := util.RenderTemplate(tmpl, struct {
					Replicas  int
					Namespace string
				}{
					Replicas:  replicas,
					Namespace: fmt.Sprintf("service-%d", i),
				})
				if err != nil {
					panic(err)
				}

				// Loop through all yaml documents
				for _, doc := range docs {
					if err := util.ApplyYaml(client, doc); err != nil {
						panic(err)
					}
				}
			}
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
