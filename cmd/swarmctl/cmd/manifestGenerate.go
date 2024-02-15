package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import "github.com/spf13/cobra"

//-----------------------------------------------------------------------------
// manifestGenerateCmd
//-----------------------------------------------------------------------------

var manifestGenerateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generates a manifest and outputs it.",
	Aliases: []string{"g"},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	// Add the command to the parent
	manifestCmd.AddCommand(manifestGenerateCmd)

	// --replicas flag
	manifestGenerateCmd.PersistentFlags().IntVar(&replicas, "replicas", 1, "Number of replicas to deploy.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("replicas", replicasCompletion); err != nil {
		panic(err)
	}

	// --node-selector flag
	manifestGenerateCmd.PersistentFlags().StringVar(&nodeSelector, "node-selector", "", "Node selector to use for deployment.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("node-selector", nodeSelectorCompletion); err != nil {
		panic(err)
	}

	// --image-tag flag
	manifestGenerateCmd.PersistentFlags().StringVar(&imageTag, "image-tag", "", "Image tag to use for deployment.")
	if err := manifestGenerateCmd.RegisterFlagCompletionFunc("image-tag", imageTagCompletion); err != nil {
		panic(err)
	}
}
