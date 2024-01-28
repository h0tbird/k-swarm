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
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

type context struct {
	config *rest.Config
	mapGV  map[string]*metav1.APIResourceList
}

var (
	Assets         embed.FS
	contexts       = map[string]*context{}
	ctxRegex       string
	cpuProfile     bool
	memProfile     bool
	tracing        bool
	cpuProfileFile string
	memProfileFile string
	tracingFile    string
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

		// Get the regex matches
		matches, err := util.FilterKubeContexts(ctxRegex)
		if err != nil {
			return err
		}

		// Print
		cmd.Println("\nMatched contexts:")

		// For every match
		for _, match := range matches {

			// Print the context match
			cmd.Printf("  - %s\n", match)

			// Create a config for this context
			config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: util.HomeDir + "/.kube/config"},
				&clientcmd.ConfigOverrides{CurrentContext: match},
			).ClientConfig()
			if err != nil {
				return err
			}

			// Store the config
			contexts[match] = &context{
				config: config,
				mapGV:  map[string]*metav1.APIResourceList{},
			}
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
	contexts, err := util.ListKubeContexts()
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
