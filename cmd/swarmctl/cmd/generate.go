package cmd

import (

	// Stdlib
	"embed"
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

var Assets embed.FS

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
}
