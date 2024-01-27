package cmd

import (

	// Stdlib
	"embed"
	"strings"
	"time"

	// Community
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	Assets         embed.FS
	configs        map[string]*rest.Config
	ctxRegex       string
	cpuProfile     bool
	memProfile     bool
	cpuProfileFile string
	memProfileFile string
)

//-----------------------------------------------------------------------------
// rootCmd represents the base command when called without any subcommands
//-----------------------------------------------------------------------------

var rootCmd = &cobra.Command{
	Use:   "swarmctl",
	Short: "swarmctl controls the swarm",
}

//-----------------------------------------------------------------------------
// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
//-----------------------------------------------------------------------------

func Execute() error {
	return rootCmd.Execute()
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Profiling flags
	rootCmd.PersistentFlags().BoolVar(&cpuProfile, "cpu-profile", false, "write cpu profile to file")
	rootCmd.PersistentFlags().BoolVar(&memProfile, "mem-profile", false, "write memory profile to file")
	rootCmd.PersistentFlags().StringVar(&cpuProfileFile, "cpu-profile-file", "cpu.prof", "write cpu profile to file")
	rootCmd.PersistentFlags().StringVar(&memProfileFile, "mem-profile-file", "mem.prof", "write memory profile to file")

	// Context flag
	rootCmd.PersistentFlags().StringVar(&ctxRegex, "context", "", "regex to match the context name.")
	if err := rootCmd.RegisterFlagCompletionFunc("context", contextCompletionFunc); err != nil {
		panic(err)
	}

	// Execute the pre-run before every command Run call
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {

		// Return early if the command is a completion command
		if cmd.CalledAs() == "__complete" || strings.Contains(cmd.CommandPath(), "completion") {
			return nil
		}

		// Initialize the map
		configs = make(map[string]*rest.Config)

		// Get the contexts that match the regex
		contexts, err := util.FilterKubeContexts(ctxRegex)
		if err != nil {
			return err
		}

		// Print
		cmd.Println("\nMatched contexts:")

		// For every context
		for _, context := range contexts {

			// Print the context indented
			cmd.Printf("  - %s\n", context)

			// Create a config from the context
			config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: util.HomeDir + "/.kube/config"},
				&clientcmd.ConfigOverrides{CurrentContext: context},
			).ClientConfig()
			if err != nil {
				return err
			}

			// Store the config
			configs[context] = config
		}

		// A chance to cancel
		cmd.Println("\nStarting in 3 seconds...")
		time.Sleep(3 * time.Second)

		// Return
		return nil
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
