package cmd

import (
	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// installCmd represents the install command
//-----------------------------------------------------------------------------

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Generates a manifest and applies it.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(installCmd)
}
