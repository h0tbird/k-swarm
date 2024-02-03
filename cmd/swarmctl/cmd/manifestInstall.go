package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"errors"
	"fmt"
	"strings"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/k8sctx"
)

//-----------------------------------------------------------------------------
// manifestInstallCmd
//-----------------------------------------------------------------------------

var manifestInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Generates a manifest and applies it.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

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
		cmd.Print("\nDo you want to continue? [y/N] ")
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
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("context", contextCompletionFunc); err != nil {
		panic(err)
	}

	// --replicas flag
	manifestInstallCmd.PersistentFlags().IntVar(&replicas, "replicas", 1, "Number of replicas to deploy.")

	// --node-selector flag
	manifestInstallCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
	if err := manifestInstallCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletionFunc); err != nil {
		panic(err)
	}
}

//-----------------------------------------------------------------------------
// contextCompletionFunc
//-----------------------------------------------------------------------------

func contextCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	// Get the contexts
	contexts, err := k8sctx.List()
	if err != nil {
		panic(err)
	}

	// Filter the contexts
	var completions []string
	for _, context := range contexts {
		if strings.HasPrefix(context, toComplete) {
			completions = append(completions, context)
		}
	}

	// Return the completions
	return completions, cobra.ShellCompDirectiveNoFileComp
}
