package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"embed"
	"errors"
	"fmt"
	"strings"

	// Community
	"github.com/spf13/cobra"

	// Local
	"github.com/octoroot/k-swarm/cmd/swarmctl/pkg/k8sctx"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	Assets         embed.FS
	contexts       = map[string]*k8sctx.Context{}
	ctxRegex       string
	cpuProfile     bool
	memProfile     bool
	tracing        bool
	cpuProfileFile string
	memProfileFile string
	tracingFile    string
	nodeSelector   string
)

//-----------------------------------------------------------------------------
// rootCmd represents the base command when called without any subcommands
//-----------------------------------------------------------------------------

var rootCmd = &cobra.Command{
	Use:   "swarmctl",
	Short: "swarmctl controls the swarm",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		cmd.Println("Hello from root PersistentPreRunE")

		// Return early if the command is a completion command
		if cmd.CalledAs() == "__complete" || strings.Contains(cmd.CommandPath(), "completion") {
			return nil
		}

		// Handle profiling
		onStopProfiling = startProfiling()

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
// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
//-----------------------------------------------------------------------------

func Execute() error {
	defer stopProfiling()
	return rootCmd.Execute()
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Profiling flags
	rootCmd.PersistentFlags().BoolVar(&cpuProfile, "cpu-profile", false, "write cpu profile to file")
	rootCmd.PersistentFlags().BoolVar(&memProfile, "mem-profile", false, "write memory profile to file")
	rootCmd.PersistentFlags().BoolVar(&tracing, "tracing", false, "write trace to file")
	rootCmd.PersistentFlags().StringVar(&cpuProfileFile, "cpu-profile-file", "cpu.prof", "file for CPU profiling output")
	rootCmd.PersistentFlags().StringVar(&memProfileFile, "mem-profile-file", "mem.prof", "file for memory profiling output")
	rootCmd.PersistentFlags().StringVar(&tracingFile, "tracing-file", "trace.out", "file for tracing output")

	// Context flag
	rootCmd.PersistentFlags().StringVar(&ctxRegex, "context", "", "regex to match the context name.")
	if err := rootCmd.RegisterFlagCompletionFunc("context", contextCompletionFunc); err != nil {
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
