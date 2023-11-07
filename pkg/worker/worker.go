package worker

import (

	// Stdlib
	"context"
	"sync"

	// Community
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"

	// Internal
	"github.com/octoroot/swarm/pkg/common"
)

//-----------------------------------------------------------------------------
// Start starts the worker
//-----------------------------------------------------------------------------

func Start(ctx context.Context, wg *sync.WaitGroup, flags *common.FlagPack) {

	defer wg.Done()

	//--------------------------
	// Server
	//--------------------------

	// Setup the router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Routes
	router.GET("/data", getData)

	// Start the server
	endless.ListenAndServe(flags.WorkerAddr, router)

	//--------------------------
	// Client
	//--------------------------

	// TODO: Implement the client
}

//-----------------------------------------------------------------------------
// getData
//-----------------------------------------------------------------------------

func getData(c *gin.Context) {

	// TODO: Report real data.
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
