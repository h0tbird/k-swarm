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
// manifestInstallWorkerCmd
//-------------------------------------------------------------------------

var manifestInstallWorkerCmd = &cobra.Command{
	Use:     "worker <start:end>",
	Short:   "Installs worker manifests.",
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
			panic(err)
		}

		// Loop through all contexts
		for name, context := range contexts {

			// Print the context
			fmt.Printf("\n%s\n", name)

			// Loop trough all services
			for i := start; i <= end; i++ {

				fmt.Printf("\n")

				// Render the template
				docs, err := util.RenderTemplate(tmpl, struct {
					Replicas     int
					Namespace    string
					NodeSelector string
				}{
					Replicas:     replicas,
					Namespace:    fmt.Sprintf("service-%d", i),
					NodeSelector: nodeSelector,
				})
				if err != nil {
					panic(err)
				}

				// Loop through all yaml documents
				for _, doc := range docs {
					if err := context.ApplyYaml(doc); err != nil {
						panic(err)
					}
				}
			}
		}

		// Return
		return nil
	},
}

//-------------------------------------------------------------------------
// init
//-------------------------------------------------------------------------

func init() {

	// Add command to rootCmd installCmd
	rootCmd.AddCommand(manifestInstallWorkerCmd)
	manifestInstallCmd.AddCommand(manifestInstallWorkerCmd)
}
