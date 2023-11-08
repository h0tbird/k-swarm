package common

import "time"

// FlagPack
type FlagPack struct {
	MetricsAddr          string
	EnableLeaderElection bool
	SyncPeriod           time.Duration
	ProbeAddr            string

	// Informer flags
	EnableInformer   bool
	InformerBindAddr string

	// Worker flags
	EnableWorker          bool
	WorkerBindAddr        string
	InformerPollInterval  time.Duration
	WorkerRequestInterval time.Duration
	InformerURL           string
}
