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
	ctrl "sigs.k8s.io/controller-runtime"

	// Internal
	"github.com/octoroot/swarm/pkg/common"
)

//-----------------------------------------------------------------------------
// Global variables
//-----------------------------------------------------------------------------

var (
	serviceList = []string{}
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
	endless.ListenAndServe(flags.WorkerBindAddr, router)
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

	// Get the service list from the informer
	go pollServiceList(ctx, flags, &serviceList)

	for {
		select {
		case <-ctx.Done():
			log.Println("worker client context done")
			return
		default:
			ctrl.Log.WithName("worker").Info("sending a request", "service", "TODO")
			time.Sleep(flags.WorkerRequestInterval)
		}
	}
}

//-----------------------------------------------------------------------------
// pollServiceList polls the service list from the informer
//-----------------------------------------------------------------------------

func pollServiceList(ctx context.Context, flags *common.FlagPack, serviceList *[]string) {

	// Setup a ticker
	ticker := time.NewTicker(flags.InformerPollInterval)
	defer ticker.Stop()

	// Loop
	for {
		select {
		case <-ticker.C:
			ctrl.Log.WithName("worker").Info("polling service list", "url", flags.InformerURL+"/services")
		case <-ctx.Done():
			log.Println("worker client context done")
			return
		}
	}
}
