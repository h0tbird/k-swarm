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

	// Worker flags
	EnableWorker          bool
	WorkerBindAddr        string
	WorkerRequestInterval time.Duration
	WorkerLogResponses    bool

	// Memberlist flags
	ServiceName             string
	MemberlistBindAddr      string
	MemberlistAdvertiseAddr string
	MemberlistJoinDNS       string
	MemberlistRejoinPeriod  time.Duration
}
