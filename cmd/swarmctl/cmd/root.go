package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"embed"
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
	Use:   "swarmctl",
	Short: "swarmctl controls the swarm",
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
}
