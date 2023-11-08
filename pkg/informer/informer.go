package informer

import (

	// Stdlib
	"context"
	"os"
	"sync"

	// Community
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

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

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
	if err := mgr.Add(newInformer(commChan)); err != nil {
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
}

//-----------------------------------------------------------------------------
// newInformer returns a new informer runnable
//-----------------------------------------------------------------------------

func newInformer(commChan chan []string) Informer {
	return Informer{commChan: commChan}
}

//-----------------------------------------------------------------------------
// Start starts the informer runnable
//-----------------------------------------------------------------------------

func (i Informer) Start(ctx context.Context) error {

	setupLog.Info("starting informer runnable")

	// Read from the channel and print the services
	for {
		select {
		case services := <-i.commChan:
			setupLog.Info("services", "services", services)
		case <-ctx.Done():
			setupLog.Info("stopping informer runnable")
			return nil
		}
	}
}
