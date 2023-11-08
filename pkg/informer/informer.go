package informer

import (

	// Stdlib
	"context"
	"os"
	"sync"

	// Community
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	// Internal
	"github.com/octoroot/swarm/internal/controller"
	"github.com/octoroot/swarm/pkg/common"
)

//-----------------------------------------------------------------------------
// Global variables
//-----------------------------------------------------------------------------

var (
	scheme      = runtime.NewScheme()
	setupLog    = ctrl.Log.WithName("setup")
	serviceList = []string{}
)

//-----------------------------------------------------------------------------
// init
//-----------------------------------------------------------------------------

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

//-----------------------------------------------------------------------------
// Start starts the informer
//-----------------------------------------------------------------------------

func Start(ctx context.Context, wg *sync.WaitGroup, flags *common.FlagPack) {

	defer wg.Done()

	// Initializes a new controller manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: flags.MetricsAddr},
		HealthProbeBindAddress: flags.ProbeAddr,
		LeaderElection:         flags.EnableLeaderElection,
		LeaderElectionID:       "bb4dbf8a.github.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// controller --> runnable communication channel
	commChan := make(chan []string)

	//-------------------------
	// Register the controller
	//-------------------------

	// Register the swarm controller
	if err = (&controller.ServiceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		CommChan: commChan,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "swarm")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	//-----------------------
	// Register the runnable
	//-----------------------

	// Register the informer runnable
	if err := mgr.Add(newInformer(commChan, flags)); err != nil {
		setupLog.Error(err, "unable to register informer")
		os.Exit(1)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	// Add ready checks
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	//-------------------
	// Start the manager
	//-------------------

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
// Informer implements the runnable interface
//-----------------------------------------------------------------------------

type Informer struct {
	commChan chan []string
	flags    *common.FlagPack
}

//-----------------------------------------------------------------------------
// newInformer returns a new informer runnable
//-----------------------------------------------------------------------------

func newInformer(commChan chan []string, flags *common.FlagPack) Informer {
	return Informer{
		commChan: commChan,
		flags:    flags,
	}
}

//-----------------------------------------------------------------------------
// Start starts the informer runnable
//-----------------------------------------------------------------------------

func (i Informer) Start(ctx context.Context) error {

	setupLog.Info("starting informer runnable")

	// Retrieve the services from the comm channel
	go func() {
		for {
			select {
			case serviceList = <-i.commChan:
				ctrl.Log.WithName("informer").Info("new update", "services", serviceList)
			case <-ctx.Done():
				setupLog.Info("stopping informer runnable")
				return
			}
		}
	}()

	// Setup the router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	// Routes
	router.GET("/services", getServices)

	// Start the server
	endless.ListenAndServe(i.flags.InformerBindAddr, router)

	// Return no error
	return nil
}

//-----------------------------------------------------------------------------
// getServices
//-----------------------------------------------------------------------------

func getServices(c *gin.Context) {
	c.JSON(200, gin.H{
		"services": serviceList,
	})
}
