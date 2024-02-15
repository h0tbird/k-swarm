package main

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	//Stdlib
	"embed"
	"os"

	// Internal
	"github.com/octoroot/k-swarm/cmd/swarmctl/cmd"
)

//go:embed assets/*
var assets embed.FS

//-----------------------------------------------------------------------------
// main
//-----------------------------------------------------------------------------

func main() {
	cmd.Assets = assets
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
