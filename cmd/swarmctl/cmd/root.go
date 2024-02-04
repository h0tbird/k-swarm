package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"embed"
	"errors"
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
	stdout         bool
	cpuProfileFile string
	memProfileFile string
	tracingFile    string
	nodeSelector   string
	replicas       int
)

//-----------------------------------------------------------------------------
// rootCmd represents the base command when called without any subcommands
//-----------------------------------------------------------------------------

var rootCmd = &cobra.Command{
	Version: "0.0.1",
	Use:     "swarmctl",
	Short:   "swarmctl controls the swarm",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		// Return early if the command is a completion command
		if cmd.CalledAs() == "__complete" || strings.Contains(cmd.CommandPath(), "completion") {
			return nil
		}

		// Handle profiling
		onStopProfiling = startProfiling()

		// Return
		return nil
	},
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
}

//-----------------------------------------------------------------------------
// This is called by main.main()
//-----------------------------------------------------------------------------

func Execute() error {
	defer stopProfiling()
	return rootCmd.Execute()
}

//-----------------------------------------------------------------------------
// context
//-----------------------------------------------------------------------------

// contextCompletion
func contextCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

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

// contextIsValid
func contextIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// replicas
//-----------------------------------------------------------------------------

// replicasCompletion
func replicasCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"1"}, cobra.ShellCompDirectiveNoFileComp
}

// replicasIsValid
func replicasIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// nodeSelector
//-----------------------------------------------------------------------------

// nodeSelectorCompletion
func nodeSelectorCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"{key1: value1, key2: value2}"}, cobra.ShellCompDirectiveNoFileComp
}

// nodeSelectorIsValid
func nodeSelectorIsValid() bool {
	return true
}

//-----------------------------------------------------------------------------
// validateFlags
//-----------------------------------------------------------------------------

func validateFlags(cmd *cobra.Command, args []string) error {

	if cmd.Flags().Changed("context") {
		if valid := contextIsValid(); !valid {
			return errors.New("invalid context")
		}
	}

	if cmd.Flags().Changed("replicas") {
		if valid := replicasIsValid(); !valid {
			return errors.New("invalid replicas")
		}
	}

	if cmd.Flags().Changed("node-selector") {
		if valid := nodeSelectorIsValid(); !valid {
			return errors.New("invalid node-selector")
		}
	}

	// Return
	return nil
}
