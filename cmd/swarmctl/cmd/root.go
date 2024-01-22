package cmd

import (

	// Stdlib
	"embed"
	"os"
	"path/filepath"
	"strings"

	// Community
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var Assets embed.FS

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

	// Define the flags
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	if rootCmd.PersistentFlags().String("context", "", "Regex to match the context name.") != nil {
		if err := rootCmd.RegisterFlagCompletionFunc("context", contextCompletionFunc); err != nil {
			panic(err)
		}
	}
}

//-----------------------------------------------------------------------------
// contextCompletionFunc
//-----------------------------------------------------------------------------

func contextCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	// Get the user's home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Load the kubeconfig file.
	config, err := clientcmd.LoadFromFile(filepath.Join(homeDir, ".kube", "config"))
	if err != nil {
		panic(err)
	}

	// Get the contexts from the config.
	contexts := config.Contexts

	// Filter the contexts.
	var completions []string
	for context := range contexts {
		if strings.HasPrefix(context, toComplete) {
			completions = append(completions, context)
		}
	}

	// Return the completions.
	return completions, cobra.ShellCompDirectiveNoFileComp
}
