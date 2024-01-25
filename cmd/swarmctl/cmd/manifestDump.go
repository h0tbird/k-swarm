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
// dumpCmd
//-----------------------------------------------------------------------------

var dumpCmd = &cobra.Command{
	Use:       "dump [informer|worker]",
	Short:     "Dumps templates to ~/.swarmctl or stdout.",
	ValidArgs: []string{"informer", "worker"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {

		component := args[0]
		stdout, _ := cmd.Flags().GetBool("stdout")

		// Open the file from the embedded file system
		fileData, err := Assets.ReadFile(fmt.Sprintf("assets/%s.goyaml", component))
		if err != nil {
			return fmt.Errorf("error reading file from embedded FS: %w", err)
		}

		// Write the contents to stdout
		if stdout {
			_, err = io.Copy(os.Stdout, bytes.NewReader(fileData))
			if err != nil {
				return fmt.Errorf("error writing file data to stdout: %w", err)
			}
			return nil
		}

		// Get the user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting user's home directory: %w", err)
		}

		// Create ~/.swarmctl if it doesn't exist
		if err := os.MkdirAll(homeDir+"/.swarmctl", 0755); err != nil {
			return fmt.Errorf("error creating ~/.swarmctl: %w", err)
		}

		// Write the contents to ~/.swarmctl/<component>.goyaml
		if err := os.WriteFile(homeDir+"/.swarmctl/"+component+".goyaml", fileData, 0644); err != nil {
			return fmt.Errorf("error writing file data to ~/.swarmctl/%s.goyaml: %w", component, err)
		}

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
