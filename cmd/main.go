package main

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"flag"
	"sync"
	"time"

	// Community
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// Internal
	"github.com/h0tbird/k-swarm/pkg/common"
	"github.com/h0tbird/k-swarm/pkg/informer"
	"github.com/h0tbird/k-swarm/pkg/worker"
	//+kubebuilder:scaffold:imports
)

//-----------------------------------------------------------------------------
// initFlags initializes the command line flags.
//-----------------------------------------------------------------------------

func initFlags(fs *pflag.FlagSet) *common.FlagPack {

	flags := &common.FlagPack{}

	//--------------
	// Common flags
	//--------------

	fs.StringVar(
		&flags.MetricsAddr,
		"metrics-bind-address",
		":8080",
		"The address the metric endpoint binds to.")

	flag.BoolVar(
		&flags.SecureMetrics,
		"metrics-secure",
		true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")

	flag.StringVar(
		&flags.MetricsCertPath,
		"metrics-cert-path",
		"",
		"The directory that contains the metrics server certificate.")

	flag.StringVar(
		&flags.MetricsCertName,
		"metrics-cert-name",
		"tls.crt",
		"The name of the metrics server certificate file.")

	flag.StringVar(
		&flags.MetricsCertKey,
		"metrics-cert-key",
		"tls.key",
		"The name of the metrics server key file.")

	fs.StringVar(
		&flags.ProbeAddr,
		"health-probe-bind-address",
		":8081",
		"The address the probe endpoint binds to.")

	fs.BoolVar(
		&flags.EnableLeaderElection,
		"leader-elect",
		false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	fs.DurationVar(
		&flags.SyncPeriod,
		"sync-period",
		10*time.Hour,
		"The minimum interval at which watched resources are reconciled.")

	flag.BoolVar(
		&flags.EnableHTTP2,
		"enable-http2",
		false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	//----------------
	// Informer flags
	//----------------

	fs.BoolVar(
		&flags.EnableInformer,
		"enable-informer",
		false,
		"Enable the informer.")

	fs.StringVar(
		&flags.InformerBindAddr,
		"informer-bind-address", ":8083",
		"The address the informer binds to.")

	//--------------
	// Worker flags
	//--------------

	fs.BoolVar(
		&flags.EnableWorker,
		"enable-worker",
		false,
		"Enable the worker.")

	fs.StringVar(
		&flags.WorkerBindAddr,
		"worker-bind-address", ":8082",
		"The address the worker binds to.")

	fs.StringVar(
		&flags.InformerURL,
		"informer-url",
		"http://localhost:8083",
		"The URL of the informer.")

	fs.DurationVar(
		&flags.InformerPollInterval,
		"informer-poll-interval",
		10*time.Second,
		"The interval at which the worker polls the informer.")

	fs.DurationVar(
		&flags.WorkerRequestInterval,
		"worker-request-interval",
		2*time.Second,
		"The interval at which the worker sends requests.")

	return flags
}

//-----------------------------------------------------------------------------
// main
//-----------------------------------------------------------------------------

func main() {

	zapOpts := zap.Options{Development: true}
	var wg sync.WaitGroup

	// Handle flags
	flags := initFlags(pflag.CommandLine)
	zapOpts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// Logger setup
	log := zap.New(zap.UseFlagOptions(&zapOpts))
	ctrl.SetLogger(log)
	ctrl.Log.WithName("main").Info("Starting")

	// Setup a common context
	ctx := ctrl.SetupSignalHandler()

	// Run as an informer (gin web framework)
	if flags.EnableInformer {
		wg.Add(1)
		ctrl.Log.WithName("main").Info("Starting informer")
		go informer.Start(ctx, &wg, flags)
	}

	// Run as a worker (controller-runtime)
	if flags.EnableWorker {
		wg.Add(1)
		ctrl.Log.WithName("main").Info("Starting worker")
		go worker.Start(ctx, &wg, flags)
	}

	// Wait
	wg.Wait()
	ctrl.Log.WithName("main").Info("Shutting down")
}
