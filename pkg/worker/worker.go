package worker

import (

	// Stdlib
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	// Community
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	ctrl "sigs.k8s.io/controller-runtime"

	// Internal
	"github.com/h0tbird/k-swarm/pkg/common"
)

//-----------------------------------------------------------------------------
// Global variables
//-----------------------------------------------------------------------------

var log = ctrl.Log.WithName("peer")

//-----------------------------------------------------------------------------
// Start starts the worker
//-----------------------------------------------------------------------------

func Start(ctx context.Context, wg *sync.WaitGroup, flags *common.FlagPack) {

	defer wg.Done()

	// Build local node identity from downward-API env vars and flags.
	self := nodeMeta{
		Cluster:   os.Getenv("CLUSTER_NAME"),
		Namespace: os.Getenv("POD_NAMESPACE"),
		Service:   flags.ServiceName,
	}

	advertise := flags.MemberlistAdvertiseAddr
	if advertise == "" {
		advertise = os.Getenv("POD_IP")
	}

	cl, err := newCluster(ctx, log, self,
		flags.MemberlistBindAddr, advertise,
		flags.MemberlistJoinDNS, flags.MemberlistRejoinPeriod,
	)
	if err != nil {
		log.Error(err, "unable to start memberlist cluster")
		os.Exit(1)
	}

	// Worker server responds to /data and /members
	go server(flags, cl)

	// Worker client requests /data on each known peer Service
	go client(ctx, flags, cl)

	// Block until ctx is cancelled, then leave the gossip ring.
	<-ctx.Done()
	cl.Shutdown(5 * time.Second)
}

//-----------------------------------------------------------------------------
// server starts the worker HTTP server
//-----------------------------------------------------------------------------

func server(flags *common.FlagPack, cl *cluster) {

	// Setup the router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Error(err, "unable to set trusted proxies")
		os.Exit(1)
	}

	// Routes
	router.GET("/data", getData)
	router.GET("/members", func(c *gin.Context) {
		members, services := cl.Members()
		c.JSON(200, gin.H{
			"members":  members,
			"services": services,
		})
	})

	// Start the server
	if err := endless.ListenAndServe(flags.WorkerBindAddr, router); err != nil {
		log.Error(err, "unable to start worker server")
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
// getData returns this pod's identity to the caller.
//-----------------------------------------------------------------------------

func getData(c *gin.Context) {
	c.JSON(200, localPeer())
}

//-----------------------------------------------------------------------------
// localPeer returns the identity of this pod, populated from the downward API
// env vars wired up by the worker manifest.
//-----------------------------------------------------------------------------

func localPeer() peerInfo {
	return peerInfo{
		Cluster:   os.Getenv("CLUSTER_NAME"),
		Node:      os.Getenv("NODE_NAME"),
		Namespace: os.Getenv("POD_NAMESPACE"),
		Pod:       os.Getenv("POD_NAME"),
		IP:        os.Getenv("POD_IP"),
	}
}

//-----------------------------------------------------------------------------
// client loops over every Service discovered via gossip and issues GET /data.
//-----------------------------------------------------------------------------

func client(ctx context.Context, flags *common.FlagPack, cl *cluster) {

	// Bind the worker's own identity to the logger so every line is
	// self-describing as "src -> dst" when tailing logs from many pods.
	log := log.WithValues("src", localPeer())

	for {
		select {
		case <-ctx.Done():
			log.Info("client context done")
			return
		default:
			services := cl.Services()
			if len(services) == 0 {
				time.Sleep(flags.WorkerRequestInterval)
				continue
			}
			for _, service := range services {
				time.Sleep(flags.WorkerRequestInterval)
				start := time.Now()
				resp, err := http.Get(fmt.Sprintf("http://%s/data", service))
				if err != nil {
					log.Error(err, "request failed", "service", service)
					continue
				}
				durationMs := time.Since(start).Milliseconds()
				body, readErr := io.ReadAll(resp.Body)
				if cerr := resp.Body.Close(); cerr != nil {
					log.Error(cerr, "failed to close response body", "service", service)
				}
				if readErr != nil {
					log.Error(readErr, "failed to read response body", "service", service)
					continue
				}
				if !flags.WorkerLogResponses {
					continue
				}
				var dst peerInfo
				if err := json.Unmarshal(body, &dst); err != nil {
					// Fallback: log the raw body if it isn't the expected shape.
					log.Info("hop",
						"service", service,
						"http", httpInfo{Status: resp.StatusCode},
						"duration_ms", durationMs,
						"body", string(body),
					)
					continue
				}
				log.Info("hop",
					"dst", dst,
					"http", httpInfo{Status: resp.StatusCode},
					"duration_ms", durationMs,
				)
			}
		}
	}
}

//-----------------------------------------------------------------------------
// peerInfo is the identity of a worker pod, used for the "src" logger
// binding, the "dst" log field, and the JSON body returned by /data.
//-----------------------------------------------------------------------------

type peerInfo struct {
	Cluster   string `json:"cluster"`
	Node      string `json:"node"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	IP        string `json:"ip"`
}

//-----------------------------------------------------------------------------
// httpInfo groups HTTP-level fields under a single nested object in the log
// line, leaving room for future additions (method, path, ...).
//-----------------------------------------------------------------------------

type httpInfo struct {
	Status int `json:"status"`
}
