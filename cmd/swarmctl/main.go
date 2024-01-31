package main

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	//Stdlib
	"embed"
	"fmt"
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
		fmt.Println(err)
		os.Exit(1)
	}
}
