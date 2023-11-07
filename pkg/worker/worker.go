package worker

import (

	// Stdlib
	"context"
	"log"
	"sync"
	"time"

	// Internal
	"github.com/octoroot/swarm/pkg/common"
)

func Start(ctx context.Context, wg *sync.WaitGroup, flags *common.FlagPack) {

	defer wg.Done()

	for {
		log.Println("Hello from worker")
		time.Sleep(10 * time.Second)
	}
}
