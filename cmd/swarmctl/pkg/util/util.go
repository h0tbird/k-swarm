package util

import (

	// Stdlib
	"os"
	"path/filepath"
	"regexp"

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
