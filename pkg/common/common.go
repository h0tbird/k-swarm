package common

import "time"

// FlagPack
type FlagPack struct {
	MetricsAddr          string
	EnableLeaderElection bool
	EnableInformer       bool
	EnableWorker         bool
	InformerAddr         string
	WorkerAddr           string
	ProbeAddr            string
	SyncPeriod           time.Duration
}
