package cmd

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"fmt"
	"os"
	"runtime/pprof"
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

		// Create profiling file
		f, err := os.Create(cpuProfileFile)
		if err != nil {
			fmt.Println("could not create cpu profile file")
			return stop
		}

		// Start profiling
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Println("could not start cpu profiling")
			return stop
		}

		// Add function to stop cpu profiling to doOnStop list
		doOnStop = append(doOnStop, func() {
			pprof.StopCPUProfile()
			_ = f.Close()
			fmt.Println("cpu profile stopped")
		})
	}

	// Memory profiling
	if memProfile {

		fmt.Println("memory profile enabled")

		// Create profiling file
		f, err := os.Create(memProfileFile)
		if err != nil {
			fmt.Println("could not create memory profile file")
			return stop
		}

		// Add function to stop memory profiling to doOnStop list
		doOnStop = append(doOnStop, func() {
			_ = pprof.WriteHeapProfile(f)
			_ = f.Close()
			fmt.Println("memory profile stopped")
		})
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
