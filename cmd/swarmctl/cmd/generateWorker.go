package cmd

import (

	// Stdlib
	"fmt"
	"strconv"
	"strings"
	"text/template"

	// Community
	"github.com/spf13/cobra"
)

//-------------------------------------------------------------------------
// generateWorkerCmd
//-------------------------------------------------------------------------

var generateWorkerCmd = &cobra.Command{
	Use:   "worker [start:end]",
	Short: "Outputs worker manifests.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Parse embeded template using ParseFS
		tmpl := template.Must(template.ParseFS(Assets, "assets/worker.goyaml"))

		// Get the replicas flag
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

		// Loop from start to end
		for i := start; i <= end; i++ {

			// Convert the content to a string and print it
			tmpl.Execute(cmd.OutOrStdout(), struct {
				Replicas  int
				Namespace string
			}{
				Replicas:  replicas,
				Namespace: fmt.Sprintf("service-%d", i),
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
}
