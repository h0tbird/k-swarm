package cmd

import "github.com/spf13/cobra"

//-----------------------------------------------------------------------------
// generateCmd
//-----------------------------------------------------------------------------

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a manifest and outputs it.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(generateCmd)
}
