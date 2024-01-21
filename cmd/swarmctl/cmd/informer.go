package cmd

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// informerCmd represents the informer command
//-----------------------------------------------------------------------------

var informerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Generates a swarm informer install manifest and outputs to the console.",
	Run: func(cmd *cobra.Command, args []string) {

		// Read the content from the embedded file system
		content, err := Assets.ReadFile("assets/informer.yaml")
		if err != nil {
			fmt.Println("Error reading the embedded YAML file:", err)
			return
		}

		// Convert the content to a string and print it
		fmt.Println(string(content))
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	generateCmd.AddCommand(informerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// informerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// informerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
