package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"fmt"
	"sync"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	onStopProfiling func()
	profilingOnce   sync.Once
)

//-----------------------------------------------------------------------------
// startProfiling
//-----------------------------------------------------------------------------

func startProfiling() func() {

	// doOnStop is a list of functions to be called on stop
	var doOnStop []func()

	// stop calls all necessary functions to stop profiling
	stop := func() {
		for _, d := range doOnStop {
			if d != nil {
				d()
			}
		}
	}

	// CPU profiling
	if cpuProfile {
		fmt.Println("cpu profile enabled")
	}

	// Memory profiling
	if memProfile {
		fmt.Println("memory profile enabled")
	}

	// Return
	return stop
}

//-----------------------------------------------------------------------------
// stopProfiling
//-----------------------------------------------------------------------------

func stopProfiling() {
	if onStopProfiling != nil {
		profilingOnce.Do(onStopProfiling)
	}
}
