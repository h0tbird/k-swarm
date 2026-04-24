package main

//-----------------------------------------------------------------------------
// Imports
//-----------------------------------------------------------------------------

import (

	// Stdlib
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	// Community
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// Internal
	"github.com/h0tbird/k-swarm/pkg/common"
	"github.com/h0tbird/k-swarm/pkg/worker"
)

//-----------------------------------------------------------------------------
// initFlags initializes the command line flags.
//-----------------------------------------------------------------------------

func initFlags(fs *pflag.FlagSet) *common.FlagPack {

	flags := &common.FlagPack{}

	//---------------
	// Common flags
	//---------------

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

	fs.DurationVar(
		&flags.WorkerRequestInterval,
		"worker-request-interval",
		2*time.Second,
		"The interval at which the worker sends requests.")

	fs.BoolVar(
		&flags.WorkerLogResponses,
		"worker-log-responses",
		false,
		"If set, log the raw JSON response bodies received from peer workers' /data endpoint.")

	//------------------
	// Memberlist flags
	//------------------

	fs.StringVar(
		&flags.ServiceName,
		"service-name",
		"",
		"The Kubernetes Service this worker pod belongs to. Advertised to other peers via gossip. Required when --enable-worker is set.")

	fs.StringVar(
		&flags.MemberlistBindAddr,
		"memberlist-bind-address",
		":7946",
		"The address (host:port) memberlist binds to for gossip (TCP+UDP).")

	fs.StringVar(
		&flags.MemberlistAdvertiseAddr,
		"memberlist-advertise-address",
		"",
		"The address memberlist advertises to other peers. Defaults to POD_IP env var.")

	fs.StringVar(
		&flags.MemberlistJoinDNS,
		"memberlist-join-dns",
		"",
		"DNS name (typically a headless Service) whose A records list peer pod IPs used to bootstrap the gossip ring.")

	fs.DurationVar(
		&flags.MemberlistRejoinPeriod,
		"memberlist-rejoin-period",
		30*time.Second,
		"Interval at which to re-resolve --memberlist-join-dns and re-attempt joining stragglers.")

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

	// Validate worker-required flags.
	if flags.EnableWorker && flags.ServiceName == "" {
		fmt.Fprintln(os.Stderr, "--service-name is required when --enable-worker is set")
		os.Exit(2)
	}

	// Setup a common context
	ctx := ctrl.SetupSignalHandler()

	// Run as a worker
	if flags.EnableWorker {
		wg.Add(1)
		ctrl.Log.WithName("main").Info("Starting peer")
		go worker.Start(ctx, &wg, flags)
	}

	// Wait
	wg.Wait()
	ctrl.Log.WithName("main").Info("Shutting down")
}
