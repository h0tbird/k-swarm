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
	"github.com/octoroot/k-swarm/pkg/common"
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
			log.Info("client context done")
			return
		default:
			for _, service := range serviceList {
				time.Sleep(flags.WorkerRequestInterval)
				log.Info("sending a request", "service", service)
				_, err := http.Get(fmt.Sprintf("http://%s/data", service))
				if err != nil {
					log.Error(err, "failed to send request", "service", service)
					continue
				}
			}
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
			log.Info("polling service list", "url", flags.InformerURL+"/services")
			newServices, err := fetchServices(flags.InformerURL + "/services")
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

func fetchServices(url string) ([]string, error) {

	// Get the services
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code: %d", resp.StatusCode)
	}

	// Read the body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
