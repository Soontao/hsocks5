package backend

import (
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

// Dial function
type Dial func(network, addr string) (c net.Conn, err error)

// NewDialerProxyBackendDefault instance
func NewDialerProxyBackendDefault(dial Dial) ProxyBackend {
	return NewDialerProxyBackend(dial, DefaultHealthCheckEndpoint)
}

// NewDialerProxyBackend instance
func NewDialerProxyBackend(dial Dial, healthCheckEndpoint string) ProxyBackend {
	rt := &dialerProxyBackend{dial: dial, healthCheckEndpoint: healthCheckEndpoint}
	return rt
}

type dialerProxyBackend struct {
	name                string
	rtt                 time.Duration
	healthCheckEndpoint string
	dial                Dial
}

func (p *dialerProxyBackend) GetDialer() proxy.Dialer {
	return p
}

func (p *dialerProxyBackend) Dial(network, addr string) (c net.Conn, err error) {
	return p.dial(network, addr)
}

func (p *dialerProxyBackend) GetName() string {
	return p.name
}

func (p *dialerProxyBackend) IsHealth() bool {
	return p.rtt != RTTErrorHappened
}

func (p *dialerProxyBackend) HealthCheck() error {
	c := &http.Client{Transport: &http.Transport{Dial: p.dial}}
	t1 := time.Now()
	res, err := c.Head(p.healthCheckEndpoint) // HEAD without content
	if err != nil {
		p.rtt = RTTErrorHappened
		return err
	}
	if res.StatusCode != 200 {
		p.rtt = RTTErrorHappened
		return ErrHealthCheckStatusNotOk
	}
	p.rtt = time.Now().Sub(t1)
	return nil
}

func (p *dialerProxyBackend) GetPingRTT() time.Duration {
	return p.rtt
}
