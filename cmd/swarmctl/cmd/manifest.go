package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"errors"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// manifestCmd
//-----------------------------------------------------------------------------

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "The manifest command generates swarm manifests.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		// Validate the --node-selector flag
		if cmd.Flags().Changed("node-selector") {
			res, err := nodeSelectorIsValid()
			if err != nil {
				return err
			}
			if !res {
				return errors.New("invalid --node-selector flag")
			}
		}

		// All good
		return nil
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the parent
	rootCmd.AddCommand(manifestCmd)
}

//-----------------------------------------------------------------------------
// nodeSelectorCompletionFunc
//-----------------------------------------------------------------------------

func nodeSelectorCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"{key1: value1, key2: value2}"}, cobra.ShellCompDirectiveNoFileComp
}

//-----------------------------------------------------------------------------
// nodeSelectorIsValid
//-----------------------------------------------------------------------------

func nodeSelectorIsValid() (bool, error) {
	return true, nil
}
