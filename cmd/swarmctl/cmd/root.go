package cmd

import (

	// Stdlib
	"embed"
	"os"
	"strings"

	// Community
	"github.com/octoroot/swarm/cmd/swarmctl/pkg/util"
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	Assets   embed.FS
	contexts []string
	homeDir  string
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

		// Get the regex
		regex, err := cmd.Flags().GetString("context")
		if err != nil {
			return err
		}

		// Get the contexts that match the regex
		contexts, err = util.GetKubeContexts(regex)
		if err != nil {
			return err
		}

		return nil
	}

	// Get the home directory
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
}

//-----------------------------------------------------------------------------
// contextCompletionFunc
//-----------------------------------------------------------------------------

func contextCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	// Get the contexts
	contexts, err := util.GetKubeContexts("")
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
