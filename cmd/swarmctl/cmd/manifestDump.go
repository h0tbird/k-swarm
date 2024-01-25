package cmd

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// dumpCmd represents the generate command
//-----------------------------------------------------------------------------

var dumpCmd = &cobra.Command{
	Use:       "dump [informer|worker]",
	Short:     "Dumps templates to ~/.swarmctl or stdout.",
	ValidArgs: []string{"informer", "worker"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		component := args[0]

		stdout, _ := cmd.Flags().GetBool("stdout")
		if stdout {
			// Add logic to handle the dump to stdout here
			fmt.Printf("Dumping %s template to stdout...\n", component)
		} else {
			// Add logic to handle the dump to file or other output here
			fmt.Printf("Dumping %s template to stdout...\n", component)
		}
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().BoolP("stdout", "", false, "Output to stdout")
}
