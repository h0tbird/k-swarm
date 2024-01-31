/*
Copyright © 2024 GitHub, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

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
	"github.com/octoroot/k-swarm/pkg/common"
	"github.com/octoroot/k-swarm/pkg/informer"
	"github.com/octoroot/k-swarm/pkg/worker"
	//+kubebuilder:scaffold:imports
)

//-----------------------------------------------------------------------------
// initFlags initializes the command line flags.
//-----------------------------------------------------------------------------

func initFlags(fs *pflag.FlagSet) *common.FlagPack {

	flags := &common.FlagPack{}

	fs.StringVar(
		&flags.MetricsAddr,
		"metrics-bind-address",
		":8080",
		"The address the metric endpoint binds to.")

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

	zapOpts := zap.Options{}
	var wg sync.WaitGroup

	// Handle flags
	flags := initFlags(pflag.CommandLine)
	zapOpts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// Logger setup
	log := zap.New(zap.UseFlagOptions(&zapOpts))
	ctrl.SetLogger(log)

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
}
