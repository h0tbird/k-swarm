package cmd

import (

	// Stdlib
	"embed"
	"os"
	"strings"
	"time"

	// Community
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	Assets     embed.FS
	clientsets map[string]*kubernetes.Clientset
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Register the context flag completion function
	if rootCmd.PersistentFlags().String("context", "", "Regex to match the context name.") != nil {
		if err := rootCmd.RegisterFlagCompletionFunc("context", contextCompletionFunc); err != nil {
			panic(err)
		}
	}

	// Execute the pre-run before every command Run call
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {

		// Return early if the command is a completion command
		if cmd.CalledAs() == "__complete" || strings.Contains(cmd.CommandPath(), "completion") {
			return nil
		}

		// Initialize the map
		clientsets = make(map[string]*kubernetes.Clientset)

		// Get the regex
		regex, err := cmd.Flags().GetString("context")
		if err != nil {
			return err
		}

		// Get the contexts that match the regex
		contexts, err := util.FilterKubeContexts(regex)
		if err != nil {
			return err
		}

		// Print
		cmd.Println("Used contexts:")

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

			// Create a clientset from the config
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}

			// Store the clientset in the map
			clientsets[context] = clientset
		}

		// A chance to cancel
		cmd.Println("\nSleeping 2 seconds...")
		time.Sleep(2 * time.Second)

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
