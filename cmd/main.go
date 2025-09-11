package main

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"flag"
<<<<<<< HEAD
	"os"
=======
	"sync"
	"time"
>>>>>>> tmp-original-11-09-25-16-03

	// Community
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
<<<<<<< HEAD
	"sigs.k8s.io/controller-runtime/pkg/healthz"
=======
>>>>>>> tmp-original-11-09-25-16-03
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

<<<<<<< HEAD
	// Initial webhook TLS options
	webhookTLSOpts := tlsOpts
	webhookServerOptions := webhook.Options{
		TLSOpts: webhookTLSOpts,
	}
=======
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
>>>>>>> tmp-original-11-09-25-16-03

	fs.DurationVar(
		&flags.WorkerRequestInterval,
		"worker-request-interval",
		2*time.Second,
		"The interval at which the worker sends requests.")

<<<<<<< HEAD
		webhookServerOptions.CertDir = webhookCertPath
		webhookServerOptions.CertName = webhookCertName
		webhookServerOptions.KeyName = webhookCertKey
	}

	webhookServer := webhook.NewServer(webhookServerOptions)

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,
	// this setup is not recommended for production.
	//
	// TODO(user): If you enable certManager, uncomment the following lines:
	// - [METRICS-WITH-CERTS] at config/default/kustomization.yaml to generate and use certificates
	// managed by cert-manager for the metrics server.
	// - [PROMETHEUS-WITH-CERTS] at config/prometheus/kustomization.yaml for TLS certification.
	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		metricsServerOptions.CertDir = metricsCertPath
		metricsServerOptions.CertName = metricsCertName
		metricsServerOptions.KeyName = metricsCertKey
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "760bdb32.github.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := (&controller.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
=======
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

	// Run as an informer
	if flags.EnableInformer {
		wg.Add(1)
		ctrl.Log.WithName("main").Info("Starting informer")
		go informer.Start(ctx, &wg, flags)
	}

	// Run as a worker
	if flags.EnableWorker {
		wg.Add(1)
		ctrl.Log.WithName("main").Info("Starting worker")
		go worker.Start(ctx, &wg, flags)
	}

	// Wait
	wg.Wait()
	ctrl.Log.WithName("main").Info("Shutting down")
>>>>>>> tmp-original-11-09-25-16-03
}
