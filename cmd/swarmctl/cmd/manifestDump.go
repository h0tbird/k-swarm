package cmd

import "github.com/spf13/cobra"

//-----------------------------------------------------------------------------
// generateCmd represents the generate command
//-----------------------------------------------------------------------------

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps templates to ~/.swarmctl or stdout.",
}

// TODO: Use a positional argument to specify the template to dump.
// It can be either informer or worker.

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(dumpCmd)
}
