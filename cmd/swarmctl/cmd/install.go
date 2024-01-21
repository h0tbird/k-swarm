package cmd

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// installCmd represents the install command
//-----------------------------------------------------------------------------

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "The install command generates a swarm install manifest and applies it to a cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("install called")
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(installCmd)
}
