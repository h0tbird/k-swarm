package common

import "time"

// FlagPack
type FlagPack struct {
	MetricsAddr          string
	EnableLeaderElection bool
	EnableInformer       bool
	EnableWorker         bool
	ProbeAddr            string
	SyncPeriod           time.Duration
}
