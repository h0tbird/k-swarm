package util

import (

	// Stdlib
	"bytes"
	"embed"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	// Community
	"k8s.io/client-go/tools/clientcmd"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	HomeDir  string
	SwarmDir string
)

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {

	var err error

	// Get the user's home directory
	HomeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Set the swarm directory
	SwarmDir = filepath.Join(HomeDir, ".swarmctl")
}

//-----------------------------------------------------------------------------
// ListKubeContexts
//-----------------------------------------------------------------------------

func ListKubeContexts() ([]string, error) {

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(filepath.Join(HomeDir, ".kube", "config"))
	if err != nil {
		return nil, err
	}

	// Iterate over the contexts
	var contexts []string
	for context := range config.Contexts {
		contexts = append(contexts, context)
	}

	// Return the contexts.
	return contexts, nil
}

//-------------------------------------------------------------------------
// FilterKubeContexts
//-------------------------------------------------------------------------

func FilterKubeContexts(regex string) ([]string, error) {

	// Variables
	var matchingContexts []string

	// Return empty list if regex is empty
	if regex == "" {
		return matchingContexts, nil
	}

	// List the contexts
	contexts, err := ListKubeContexts()
	if err != nil {
		return nil, err
	}

	// Iterate over the contexts
	for _, context := range contexts {

		// If the context matches the regex
		if match, err := regexp.MatchString(regex, context); match && err == nil {
			matchingContexts = append(matchingContexts, context)
		} else if err != nil {
			return nil, err
		}
	}

	// Return the matching contexts.
	return matchingContexts, nil
}

//-----------------------------------------------------------------------------
// ParseTemplate
//-----------------------------------------------------------------------------

func ParseTemplate(assets embed.FS, component string) (*template.Template, error) {

	// Variables
	var tmpl *template.Template
	var err error

	// Check for the ~/.swarmctl/<component>.goyaml file
	_, err = os.Stat(SwarmDir + "/" + component + ".goyaml")

	// Use the embedded template
	if os.IsNotExist(err) {
		tmpl, err = template.ParseFS(assets, "assets/"+component+".goyaml")
		if err != nil {
			return nil, err
		}
	} else if err == nil {
		tmpl, err = template.ParseFiles(SwarmDir + "/" + component + ".goyaml")
		if err != nil {
			return nil, err
		}
	}

	// Return
	return tmpl, nil
}

//-----------------------------------------------------------------------------
// RenderTemplate
//-----------------------------------------------------------------------------

func RenderTemplate(tmpl *template.Template, data any) ([]string, error) {

	// Render the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	// Split the YAML into docs
	var docs []string
	for _, doc := range strings.Split(buf.String(), "---") {
		if strings.TrimSpace(doc) != "" {
			docs = append(docs, doc)
		}
	}

	// Return
	return docs, nil
}
