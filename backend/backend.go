package backend

import (
	"time"

	"golang.org/x/net/proxy"
)

// ProxyBackend interface
type ProxyBackend interface {
	GetDialer() proxy.Dialer
	GetName() string
	// check dialer health, return error if anything wrong
	// if success, record the RTT time
	HealthCheck() error
	GetPingRTT() time.Duration
	IsHealth() bool
}
