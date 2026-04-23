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

var (
	serviceList = []string{}
	log         = ctrl.Log.WithName("worker")
)

//-----------------------------------------------------------------------------
// Structs
//-----------------------------------------------------------------------------

type InformerData struct {
	Services []string `json:"services"`
}

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
	router := gin.New()
	router.Use(gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Error(err, "unable to set trusted proxies")
		os.Exit(1)
	}

	// Routes
	router.GET("/data", getData)

	// Start the server
	if err := endless.ListenAndServe(flags.WorkerBindAddr, router); err != nil {
		log.Error(err, "unable to start worker server")
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
// getData
//-----------------------------------------------------------------------------

func getData(c *gin.Context) {
	c.JSON(200, peerInfo{
		ClusterName:  os.Getenv("CLUSTER_NAME"),
		NodeName:     os.Getenv("NODE_NAME"),
		PodName:      os.Getenv("POD_NAME"),
		PodNamespace: os.Getenv("POD_NAMESPACE"),
		PodIP:        os.Getenv("POD_IP"),
	})
}

//-----------------------------------------------------------------------------
// client starts the worker client
//-----------------------------------------------------------------------------

func client(ctx context.Context, flags *common.FlagPack) {

	// Bind the worker's own identity to the logger so every line is
	// self-describing as "src -> dst" when tailing logs from many pods.
	log := log.WithValues("src", peerInfo{
		ClusterName:  os.Getenv("CLUSTER_NAME"),
		NodeName:     os.Getenv("NODE_NAME"),
		PodName:      os.Getenv("POD_NAME"),
		PodNamespace: os.Getenv("POD_NAMESPACE"),
		PodIP:        os.Getenv("POD_IP"),
	})

	// Get the service list from the informer
	go pollServiceList(ctx, flags, &serviceList)

	for {
		select {
		case <-ctx.Done():
			log.Info("client context done")
			return
		default:
			for _, service := range serviceList {
				time.Sleep(flags.WorkerRequestInterval)
				start := time.Now()
				resp, err := http.Get(fmt.Sprintf("http://%s/data", service))
				if err != nil {
					log.Error(err, "request failed", "dst", service)
					continue
				}
				ms := time.Since(start).Milliseconds()
				if flags.WorkerLogResponses {
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Error(err, "failed to read response body", "dst", service)
					} else {
						var peer peerInfo
						if err := json.Unmarshal(body, &peer); err != nil {
							// Fallback: log the raw body if it isn't the expected shape.
							log.Info("hop", "dst", service, "status", resp.StatusCode, "ms", ms, "body", string(body))
						} else {
							log.Info("hop", "dst", service, "status", resp.StatusCode, "ms", ms, "peer", peer)
						}
					}
				} else {
					_, _ = io.Copy(io.Discard, resp.Body)
					log.Info("hop", "dst", service, "status", resp.StatusCode, "ms", ms)
				}
				if err := resp.Body.Close(); err != nil {
					log.Error(err, "failed to close response body", "dst", service)
				}
			}
		}
	}
}

//-----------------------------------------------------------------------------
// peerInfo is the identity of a worker pod, both for the local "src" logger
// binding and for decoding the JSON body returned by /data on the remote end.
// Short JSON tags keep log lines compact when tailing with `kubectl logs`.
//-----------------------------------------------------------------------------

type peerInfo struct {
	ClusterName  string `json:"cluster"`
	NodeName     string `json:"node"`
	PodName      string `json:"pod"`
	PodNamespace string `json:"ns"`
	PodIP        string `json:"ip"`
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
			log.Info("polling service list", "url", flags.InformerURL+"/services")
			newServices, err := fetchServices(flags.InformerURL+"/services", flags.WorkerLogResponses)
			if err != nil {
				log.Error(err, "failed to fetch services")
				continue
			}
			*serviceList = newServices
		case <-ctx.Done():
			log.Info("client context done")
			return
		}
	}
}

//-----------------------------------------------------------------------------
// fetchServices fetches the services from the informer
//-----------------------------------------------------------------------------

func fetchServices(url string, logBody bool) ([]string, error) {

	// Get the services
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Defer closing the response body
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(err, "failed to close response body")
		}
	}()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	// Read the body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Optionally log the raw response body
	if logBody {
		log.Info("services response", "url", url, "body", string(bodyBytes))
	}

	// Unmarshal the body
	var data InformerData
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, err
	}

	// Filter out any services with empty names
	var services []string
	for _, service := range data.Services {
		if service != "" {
			services = append(services, service)
		}
	}

	// Return the list
	return services, nil
}
