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
	Use:          "worker <start:end>",
	Short:        "Installs worker manifests.",
	SilenceUsage: true,
	Example: `
  # Install the workers 1 to 1 to the current context
  swarmctl manifest install worker 1:1

  # Same using command aliases
  swarmctl m i w 1:1

  # Same using a shoret command chain
  swarmctl worker 1:1

  # Same using a short command chain with aliases
  swarmctl w 1:1

  # Install the workers 1 to 1 to a specific context
  swarmctl w 1:1 --context my-context

  # Install the workers 1 to 1 to all contexts that match a regex
  swarmctl w 1:1 --context 'my-.*'

  # Install the workers 1 to 1 to all contexts that match a regex and set the replicas
  swarmctl w 1:1 --context 'my-.*' --replicas 3

  # Install the workers 1 to 1 to all contexts that match a regex and set the node selector
  swarmctl w 1:1 --context 'my-.*' --node-selector '{key1: value1, key2: value2}'
`,
	Aliases: []string{"w"},
	Args:    cobra.ExactArgs(1),
	PreRunE: validateFlags,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Set the error prefix
		cmd.SetErrPrefix("\nError: ")

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
					Version      string
				}{
					Replicas:     replicas,
					Namespace:    fmt.Sprintf("service-%d", i),
					NodeSelector: nodeSelector,
					Version:      version,
				})
				if err != nil {
					return err
				}

				// Loop through all yaml documents
				for _, doc := range docs {
					if err := context.ApplyYaml(doc); err != nil {
						return err
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
