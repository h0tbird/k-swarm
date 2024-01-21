package cmd

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

//-------------------------------------------------------------------------
// workerCmd represents the worker command
//-------------------------------------------------------------------------

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Generates a swarm worker install manifest and outputs to the console.",
	Run: func(cmd *cobra.Command, args []string) {

		// Read the content from the embedded file system
		content, err := Assets.ReadFile("assets/worker.yaml")
		if err != nil {
			fmt.Println("Error reading the embedded YAML file:", err)
			return
		}

		// Convert the content to a string and print it
		fmt.Println(string(content))
	},
}

//-------------------------------------------------------------------------
// init
//-------------------------------------------------------------------------

func init() {
	generateCmd.AddCommand(workerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// workerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// workerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
