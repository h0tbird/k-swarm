package cmd

import (

	// Stdlib
	"bytes"
	"fmt"
	"io"
	"os"

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
	RunE: func(cmd *cobra.Command, args []string) error {

		component := args[0]
		stdout, _ := cmd.Flags().GetBool("stdout")

		if stdout {
			if err := dumpFileToStdout(fmt.Sprintf("assets/%s.goyaml", component)); err != nil {
				return fmt.Errorf("error dumping %s template to stdout: %w", component, err)
			}
			return nil
		}

		// TODO: Add logic to dump to ~/.swarmctl

		return nil
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().BoolP("stdout", "", false, "Output to stdout")
}

//-----------------------------------------------------------------------------
// dumpFileToStdout
//-----------------------------------------------------------------------------

func dumpFileToStdout(filename string) error {

	// Open the file from the embedded file system
	fileData, err := Assets.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file from embedded FS: %w", err)
	}

	// Write the contents to stdout
	_, err = io.Copy(os.Stdout, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("error writing file data to stdout: %w", err)
	}

	return nil
}
