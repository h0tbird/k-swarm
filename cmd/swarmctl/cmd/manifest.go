package cmd

import (
	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// manifestCmd
//-----------------------------------------------------------------------------

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "The manifest command generates swarm manifests.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(manifestCmd)
}
