package cmd

import (

	// Stdlib
	"os"
	"text/template"

	// Community
	"github.com/spf13/cobra"
)

//-----------------------------------------------------------------------------
// generateCmd
//-----------------------------------------------------------------------------

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a manifest and outputs it.",
}

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	manifestCmd.AddCommand(generateCmd)
}

//-----------------------------------------------------------------------------
// parseTemplate
//-----------------------------------------------------------------------------

func parseTemplate(component string) *template.Template {

	// Check if the file exists
	_, err := os.Stat(homeDir + "/.swarmctl/" + component + ".goyaml")

	// If it doesn't exist, use the embedded template
	if os.IsNotExist(err) {
		return template.Must(template.ParseFS(Assets, "assets/"+component+".goyaml"))
	}

	// Otherwise, use the file
	return template.Must(template.ParseFiles(homeDir + "/.swarmctl/" + component + ".goyaml"))
}
