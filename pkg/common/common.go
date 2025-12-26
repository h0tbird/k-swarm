package common

import "time"

// FlagPack
type FlagPack struct {

	// Common flags
	MetricsAddr          string
	MetricsCertPath      string
	MetricsCertName      string
	MetricsCertKey       string
	SecureMetrics        bool
	EnableLeaderElection bool
	SyncPeriod           time.Duration
	ProbeAddr            string
	EnableHTTP2          bool

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
