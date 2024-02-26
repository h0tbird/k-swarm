package profiling

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"fmt"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"sync"
)

//-----------------------------------------------------------------------------
// Globals
//-----------------------------------------------------------------------------

var (
	OnStop         func()
	once           sync.Once
	CPUProfile     bool
	CPUProfileFile string
	MemProfile     bool
	MemProfileFile string
	Tracing        bool
	TracingFile    string
)

//-----------------------------------------------------------------------------
// Start
//-----------------------------------------------------------------------------

func Start() func() {

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

	//---------------
	// CPU profiling
	//---------------

	if CPUProfile {

		fmt.Println("cpu profile enabled")

		// Create profiling file
		f, err := os.Create(CPUProfileFile)
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
			fmt.Println("\ncpu profile stopped")
		})
	}

	//------------------
	// Memory profiling
	//------------------

	if MemProfile {

		fmt.Println("memory profile enabled")

		// Create profiling file
		f, err := os.Create(MemProfileFile)
		if err != nil {
			fmt.Println("could not create memory profile file")
			return stop
		}

		// Start profiling
		err = pprof.WriteHeapProfile(f)
		if err != nil {
			fmt.Println("could not start memory profiling")
			return stop
		}

		// Add function to stop memory profiling to doOnStop list
		doOnStop = append(doOnStop, func() {
			_ = pprof.WriteHeapProfile(f)
			_ = f.Close()
			fmt.Println("\nmemory profile stopped")
		})
	}

	//---------
	// Tracing
	//---------

	if Tracing {

		fmt.Println("tracing enabled")

		// Create tracing file
		f, err := os.Create(TracingFile)
		if err != nil {
			fmt.Println("could not create tracing file")
			return stop
		}

		// Start tracing
		err = trace.Start(f)
		if err != nil {
			fmt.Println("could not start tracing")
			return stop
		}

		// Add function to stop tracing to doOnStop list
		doOnStop = append(doOnStop, func() {
			trace.Stop()
			_ = f.Close()
			fmt.Println("\ntracing stopped")
		})
	}

	// Return
	return stop
}

//-----------------------------------------------------------------------------
// Stop
//-----------------------------------------------------------------------------

func Stop() {
	if OnStop != nil {
		once.Do(OnStop)
	}
}
