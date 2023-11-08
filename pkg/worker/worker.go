package worker

import (

	// Stdlib
	"context"
	"log"
	"os"
	"sync"
	"time"

	// Community
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"

	// Internal
	"github.com/octoroot/swarm/pkg/common"
)

//-----------------------------------------------------------------------------
// Global variables
//-----------------------------------------------------------------------------

var (
	services = []string{}
)

//-----------------------------------------------------------------------------
// Start starts the worker
//-----------------------------------------------------------------------------

func Start(ctx context.Context, wg *sync.WaitGroup, flags *common.FlagPack) {

	defer wg.Done()

	// Worker server respons /data
	go server(flags)

	// Worker client requests /data
	client(ctx, flags)
}

//-----------------------------------------------------------------------------
// server starts the worker server
//-----------------------------------------------------------------------------

func server(flags *common.FlagPack) {

	// Setup the router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Routes
	router.GET("/data", getData)

	// Start the server
	endless.ListenAndServe(flags.WorkerAddr, router)
}

//-----------------------------------------------------------------------------
// getData
//-----------------------------------------------------------------------------

func getData(c *gin.Context) {
	c.JSON(200, gin.H{
		"clusterName":  os.Getenv("CLUSTER_NAME"),
		"podName":      os.Getenv("POD_NAME"),
		"podNamespace": os.Getenv("POD_NAMESPACE"),
		"podIP":        os.Getenv("POD_IP"),
		"nodeName":     os.Getenv("NODE_NAME"),
	})
}

//-----------------------------------------------------------------------------
// client starts the worker client
//-----------------------------------------------------------------------------

func client(ctx context.Context, flags *common.FlagPack) {

	// TODO: Get service list from informer

	for {
		select {
		case <-ctx.Done():
			log.Println("worker client context done")
			return
		default:
			log.Println("worker client doing something")
			time.Sleep(10 * time.Second)
		}
	}
}
