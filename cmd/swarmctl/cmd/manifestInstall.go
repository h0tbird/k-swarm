package cmd

import "github.com/spf13/cobra"

//-----------------------------------------------------------------------------
// installCmd
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
