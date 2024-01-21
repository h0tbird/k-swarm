package cmd

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// generateCmd represents the generate command
//-----------------------------------------------------------------------------

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a swarm install manifest and outputs to the console.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generate called")
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
