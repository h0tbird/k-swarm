package cmd

import (

	// Stdlib
	"fmt"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// manifestCmd represents the manifest command
//-----------------------------------------------------------------------------

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "The manifest command generates swarm manifests.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("manifest called")
	},
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	rootCmd.AddCommand(manifestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// manifestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// manifestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
