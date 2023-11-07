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
		"message": "pong",
	})
}
