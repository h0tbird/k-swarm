package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"errors"
	"fmt"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/k8sctx"
)

//-----------------------------------------------------------------------------
// manifestInstallCmd
//-----------------------------------------------------------------------------

var manifestInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Generates a manifest and applies it.",
	Aliases: []string{"i"},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		// Run the root PersistentPreRunE
		if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// Get the contexts that match the regex
		matches, err := k8sctx.Filter(ctxRegex)
		if err != nil {
			return err
		}

		// Print
		cmd.Println("\nMatched contexts:")

		// For every match
		for _, match := range matches {

			// Print the match
			cmd.Printf("  - %s\n", match)

			// Create the context
			c, err := k8sctx.New(match)
			if err != nil {
				return err
			}

			// Store the config
			contexts[match] = c
		}

		// A chance to cancel
		cmd.Print("\nProceed? (y/N) ")
		var answer string
		if _, err := fmt.Scanln(&answer); err != nil {
			return err
		}
		if answer != "y" {
			return errors.New("aborted")
		}

		// Return
		return nil
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the parent
	manifestCmd.AddCommand(manifestInstallCmd)

	// --context flag
	manifestInstallCmd.PersistentFlags().StringVar(&ctxRegex, "context", "", "regex to match the context name.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("context", contextCompletion); err != nil {
		panic(err)
	}

	// --replicas flag
	manifestInstallCmd.PersistentFlags().IntVar(&replicas, "replicas", 1, "Number of replicas to deploy.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("replicas", replicasCompletion); err != nil {
		panic(err)
	}

	// --node-selector flag
	manifestInstallCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletion); err != nil {
		panic(err)
	}

	// --image-tag flag
	manifestInstallCmd.PersistentFlags().StringVar(&imageTag, "image-tag", "", "Image tag to use for deployment.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("image-tag", imageTagCompletion); err != nil {
		panic(err)
	}
}
