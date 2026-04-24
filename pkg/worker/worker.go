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
	log         = ctrl.Log.WithName("peer")
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
// client starts the worker client
//-----------------------------------------------------------------------------

func client(ctx context.Context, flags *common.FlagPack) {

	// Bind the worker's own identity to the logger so every line is
	// self-describing as "src -> dst" when tailing logs from many pods.
	log := log.WithValues("src", localPeer())

	// Get the service list from the informer
	go pollServiceList(ctx, flags, &serviceList)

	// Loop over the service list and make requests to /data
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
