package util

import (

	// Stdlib
	"os"
	"path/filepath"
	"regexp"

	// Community
	"k8s.io/client-go/tools/clientcmd"
)

//-------------------------------------------------------------------------
// GetKubeContexts returns a list of contexts that match the given regex.
//-------------------------------------------------------------------------

func GetKubeContexts(regex string) ([]string, error) {

	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Load the kubeconfig file
	config, err := clientcmd.LoadFromFile(filepath.Join(homeDir, ".kube", "config"))
	if err != nil {
		return nil, err
	}

	// Iterate over the contexts
	var matchingContexts []string
	for context := range config.Contexts {

		// Check if the regex is empty or matches the context
		if regex == "" || context == regex {
			matchingContexts = append(matchingContexts, context)
			continue
		}

		// Check if the regex matches the context
		match, err := regexp.MatchString(regex, context)
		if err != nil {
			return nil, err
		}

		// If the regex matches the context, add it to the slice
		if match {
			matchingContexts = append(matchingContexts, context)
		}
	}

	// Return the matching contexts.
	return matchingContexts, nil
}
