package cmd

import (

	// Stdlib
	"embed"
	"os"
	"strings"

	// Community
	"github.com/spf13/cobra"
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

	// For demonstration, let's assume:
	contexts := []string{"context1", "context2", "context3"}

	var completions []string
	for _, context := range contexts {
		if strings.HasPrefix(context, toComplete) {
			completions = append(completions, context)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
