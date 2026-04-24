package worker

import (

	// Stdlib
	"context"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	// Community
	"github.com/go-logr/logr"
	"github.com/hashicorp/memberlist"
)

//-----------------------------------------------------------------------------
// nodeMeta is the per-node metadata gossiped to every other peer. Kept small
// so it fits in memberlist's NodeMeta limit (default 512 bytes).
//-----------------------------------------------------------------------------

type nodeMeta struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
}

//-----------------------------------------------------------------------------
// cluster wraps a memberlist gossip ring and aggregates the set of distinct
// Service names advertised by all known peers.
//-----------------------------------------------------------------------------

type cluster struct {
	ml       *memberlist.Memberlist
	self     nodeMeta
	selfMeta []byte

	mu       sync.RWMutex
	services map[string]int // service name -> refcount (number of pods advertising it)

	log     logr.Logger
	joinDNS string
}

//-----------------------------------------------------------------------------
// newCluster constructs and starts a memberlist node, registers delegates,
// performs the initial join, and starts a background re-join loop.
//-----------------------------------------------------------------------------

func newCluster(ctx context.Context, l logr.Logger, self nodeMeta, bindAddr, advertiseAddr, joinDNS string, rejoinPeriod time.Duration) (*cluster, error) {

	metaBytes, err := json.Marshal(self)
	if err != nil {
		return nil, fmt.Errorf("marshal self meta: %w", err)
	}
	if len(metaBytes) > memberlist.MetaMaxSize {
		return nil, fmt.Errorf("node meta exceeds %d bytes", memberlist.MetaMaxSize)
	}

	c := &cluster{
		self:     self,
		selfMeta: metaBytes,
		services: map[string]int{},
		log:      l,
		joinDNS:  joinDNS,
	}

	cfg := memberlist.DefaultLANConfig()
	cfg.Name = os.Getenv("POD_NAME")
	if cfg.Name == "" {
		cfg.Name = self.Service + "-" + advertiseAddr
	}
	bindHost, bindPort, err := splitHostPort(bindAddr, "0.0.0.0", 7946)
	if err != nil {
		return nil, fmt.Errorf("parse bind addr: %w", err)
	}
	cfg.BindAddr = bindHost
	cfg.BindPort = bindPort
	cfg.AdvertisePort = bindPort
	if advertiseAddr != "" {
		advHost, advPort, err := splitHostPort(advertiseAddr, advertiseAddr, bindPort)
		if err != nil {
			return nil, fmt.Errorf("parse advertise addr: %w", err)
		}
		cfg.AdvertiseAddr = advHost
		cfg.AdvertisePort = advPort
	}
	cfg.Delegate = (*clusterDelegate)(c)
	cfg.Events = (*clusterEvents)(c)
	cfg.Logger = stdlog.New(&logrWriter{l: l.WithName("memberlist")}, "", 0)

	ml, err := memberlist.Create(cfg)
	if err != nil {
		return nil, fmt.Errorf("create memberlist: %w", err)
	}
	c.ml = ml

	// Initial join attempt; failure to find peers is not fatal — we may be
	// the first pod in the ring. The background loop will keep trying.
	if n, err := c.tryJoin(); err != nil {
		l.Info("initial join had errors", "joined", n, "err", err.Error())
	} else {
		l.Info("initial join", "joined", n)
	}

	go c.rejoinLoop(ctx, rejoinPeriod)

	return c, nil
}

//-----------------------------------------------------------------------------
// tryJoin resolves the bootstrap DNS name to the current set of peer pod IPs,
// removes our own address, and asks memberlist to join them.
//-----------------------------------------------------------------------------

func (c *cluster) tryJoin() (int, error) {
	if c.joinDNS == "" {
		return 0, nil
	}
	addrs, err := net.LookupHost(c.joinDNS)
	if err != nil {
		return 0, fmt.Errorf("lookup %s: %w", c.joinDNS, err)
	}
	selfAddr := c.ml.LocalNode().Addr.String()
	peers := make([]string, 0, len(addrs))
	for _, a := range addrs {
		if a == selfAddr {
			continue
		}
		peers = append(peers, a)
	}
	if len(peers) == 0 {
		return 0, nil
	}
	return c.ml.Join(peers)
}

//-----------------------------------------------------------------------------
// rejoinLoop periodically re-resolves the bootstrap DNS to recover from
// transient failures and to discover late-arriving pods.
//-----------------------------------------------------------------------------

func (c *cluster) rejoinLoop(ctx context.Context, period time.Duration) {
	if period <= 0 {
		period = 30 * time.Second
	}
	t := time.NewTicker(period)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if n, err := c.tryJoin(); err != nil {
				c.log.V(1).Info("rejoin failed", "err", err.Error())
			} else if n > 0 {
				c.log.V(1).Info("rejoin", "joined", n)
			}
		}
	}
}

//-----------------------------------------------------------------------------
// Services returns a snapshot of the distinct Service names advertised by
// peers, excluding the local pod's own Service.
//-----------------------------------------------------------------------------

func (c *cluster) Services() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, 0, len(c.services))
	for s := range c.services {
		if s == "" || s == c.self.Service {
			continue
		}
		out = append(out, s)
	}
	return out
}

//-----------------------------------------------------------------------------
// Members returns a debug view of every known memberlist node and the
// derived service set. Used by the worker's /members HTTP endpoint.
//-----------------------------------------------------------------------------

type memberView struct {
	Name    string   `json:"name"`
	Addr    string   `json:"addr"`
	Meta    nodeMeta `json:"meta"`
	IsLocal bool     `json:"isLocal"`
}

func (c *cluster) Members() (members []memberView, services []string) {
	if c.ml == nil {
		return nil, nil
	}
	local := c.ml.LocalNode().Name
	for _, n := range c.ml.Members() {
		var meta nodeMeta
		_ = json.Unmarshal(n.Meta, &meta)
		members = append(members, memberView{
			Name:    n.Name,
			Addr:    net.JoinHostPort(n.Addr.String(), strconv.Itoa(int(n.Port))),
			Meta:    meta,
			IsLocal: n.Name == local,
		})
	}
	c.mu.RLock()
	for s := range c.services {
		services = append(services, s)
	}
	c.mu.RUnlock()
	return members, services
}

//-----------------------------------------------------------------------------
// Shutdown leaves the gossip ring gracefully and tears down the listeners.
//-----------------------------------------------------------------------------

func (c *cluster) Shutdown(timeout time.Duration) {
	if c.ml == nil {
		return
	}
	if err := c.ml.Leave(timeout); err != nil {
		c.log.Error(err, "memberlist leave failed")
	}
	if err := c.ml.Shutdown(); err != nil {
		c.log.Error(err, "memberlist shutdown failed")
	}
}

//-----------------------------------------------------------------------------
// addService / removeService maintain the refcounted service set.
//-----------------------------------------------------------------------------

func (c *cluster) addService(name string) (added bool) {
	if name == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	prev := c.services[name]
	c.services[name] = prev + 1
	return prev == 0
}

func (c *cluster) removeService(name string) (removed bool) {
	if name == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	prev, ok := c.services[name]
	if !ok {
		return false
	}
	if prev <= 1 {
		delete(c.services, name)
		return true
	}
	c.services[name] = prev - 1
	return false
}

//-----------------------------------------------------------------------------
// clusterDelegate implements memberlist.Delegate. Only NodeMeta is meaningful.
//-----------------------------------------------------------------------------

type clusterDelegate cluster

func (d *clusterDelegate) NodeMeta(limit int) []byte {
	if len(d.selfMeta) > limit {
		return d.selfMeta[:limit]
	}
	return d.selfMeta
}
func (d *clusterDelegate) NotifyMsg([]byte)                {}
func (d *clusterDelegate) GetBroadcasts(int, int) [][]byte { return nil }
func (d *clusterDelegate) LocalState(bool) []byte          { return nil }
func (d *clusterDelegate) MergeRemoteState([]byte, bool)   {}

//-----------------------------------------------------------------------------
// clusterEvents implements memberlist.EventDelegate, refcounting Service
// names as peers come and go.
//-----------------------------------------------------------------------------

type clusterEvents cluster

func (e *clusterEvents) NotifyJoin(n *memberlist.Node) {
	c := (*cluster)(e)
	var m nodeMeta
	_ = json.Unmarshal(n.Meta, &m)
	if c.addService(m.Service) {
		c.log.Info("service discovered", "service", m.Service, "via", n.Name)
	}
	c.log.Info("peer joined", "peer", n.Name, "addr", n.Addr.String(), "service", m.Service)
}

func (e *clusterEvents) NotifyLeave(n *memberlist.Node) {
	c := (*cluster)(e)
	var m nodeMeta
	_ = json.Unmarshal(n.Meta, &m)
	if c.removeService(m.Service) {
		c.log.Info("service gone", "service", m.Service, "via", n.Name)
	}
	c.log.Info("peer left", "peer", n.Name, "addr", n.Addr.String(), "service", m.Service)
}

func (e *clusterEvents) NotifyUpdate(n *memberlist.Node) {
	c := (*cluster)(e)
	var m nodeMeta
	_ = json.Unmarshal(n.Meta, &m)
	if c.addService(m.Service) {
		c.log.Info("service discovered", "service", m.Service, "via", n.Name)
	}
}

//-----------------------------------------------------------------------------
// logrWriter adapts an io.Writer (used by stdlib log.Logger) into logr.Info
// calls so memberlist's internal logging flows through controller-runtime.
//-----------------------------------------------------------------------------

type logrWriter struct{ l logr.Logger }

func (w *logrWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	w.l.Info(msg)
	return len(p), nil
}

//-----------------------------------------------------------------------------
// splitHostPort accepts either ":7946" or "1.2.3.4:7946" and returns the
// host and port, applying the supplied defaults when either side is absent.
//-----------------------------------------------------------------------------

func splitHostPort(addr, defaultHost string, defaultPort int) (string, int, error) {
	if addr == "" {
		return defaultHost, defaultPort, nil
	}
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		// Try treating addr as a bare host.
		return addr, defaultPort, nil
	}
	if host == "" {
		host = defaultHost
	}
	port := defaultPort
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", 0, err
		}
	}
	return host, port, nil
}
