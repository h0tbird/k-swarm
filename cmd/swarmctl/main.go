package main

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	//Stdlib
	"embed"
	"os"

	// Internal
	"github.com/h0tbird/k-swarm/cmd/swarmctl/cmd"
	"github.com/h0tbird/k-swarm/cmd/swarmctl/pkg/swarmctl"
)

//go:embed assets/*
var assets embed.FS

//-----------------------------------------------------------------------------
// main
//-----------------------------------------------------------------------------

func main() {
	swarmctl.Assets = assets
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
