package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"bytes"
	"fmt"
	"io"
	"os"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/util"
)

//-----------------------------------------------------------------------------
// dumpCmd
//-----------------------------------------------------------------------------

var dumpCmd = &cobra.Command{
	Use:       "dump [informer] [worker]",
	Short:     "Dumps templates to ~/.swarmctl or stdout.",
	ValidArgs: []string{"informer", "worker"},
	Args:      cobra.MatchAll(cobra.MaximumNArgs(2), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {

		// No args? Default to both
		if len(args) == 0 {
			args = []string{"informer", "worker"}
		}

		// Get the stdout flag
		stdout, err := cmd.Flags().GetBool("stdout")
		if err != nil {
			return fmt.Errorf("error getting stdout flag: %w", err)
		}

		// Create ~/.swarmctl
		if !stdout {
			if err := os.MkdirAll(util.SwarmDir, 0755); err != nil {
				return fmt.Errorf("error creating ~/.swarmctl: %w", err)
			}
		}

		// Loop through the components
		for _, component := range args {

			// Open the file from the embedded file system
			fileData, err := Assets.ReadFile(fmt.Sprintf("assets/%s.goyaml", component))
			if err != nil {
				return fmt.Errorf("error reading file from embedded FS: %w", err)
			}

			// Write the content to stdout
			if stdout {
				_, err = io.Copy(os.Stdout, bytes.NewReader(fileData))
				if err != nil {
					return fmt.Errorf("error writing file data to stdout: %w", err)
				}
				continue
			}

			// Write the contents to ~/.swarmctl/<component>.goyaml
			if err := os.WriteFile(util.SwarmDir+"/"+component+".goyaml", fileData, 0644); err != nil {
				return fmt.Errorf("error writing file data to ~/.swarmctl/%s.goyaml: %w", component, err)
			}

			// Print the success message
			cmd.Printf("Successfully wrote ~/.swarmctl/%s.goyaml\n", component)
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
