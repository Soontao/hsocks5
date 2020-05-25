package backend

import (
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

// DefaultHealthCheckEndpoint is google
const DefaultHealthCheckEndpoint = "https://www.google.com"

// Socks5ProxyBackendOption type
type Socks5ProxyBackendOption struct {
	Addr           string
	Name           string
	HealthEndpoint string
}

// NewSocks5ProxyBackend instance
//
// addr: a tcp address like '192.168.3.88:18080'
// name: the id of this socks5 proxy
func NewSocks5ProxyBackend(opt *Socks5ProxyBackendOption) (ProxyBackend, error) {

	dialer, err := proxy.SOCKS5("tcp", opt.Addr, nil, proxy.Direct)

	if err != nil {
		return nil, err
	}

	rt := &socks5ProxyBackend{
		dialer:              dialer,
		name:                opt.Name,
		rtt:                 RTTErrorNotCheck,
		healthCheckEndpoint: opt.HealthEndpoint,
	}

	if len(rt.healthCheckEndpoint) == 0 {
		rt.healthCheckEndpoint = DefaultHealthCheckEndpoint
	}

	return rt, nil

}

type socks5ProxyBackend struct {
	name                string
	dialer              proxy.Dialer
	rtt                 time.Duration
	healthCheckEndpoint string
}

func (p *socks5ProxyBackend) GetDialer() proxy.Dialer {
	return p.dialer
}

func (p *socks5ProxyBackend) GetName() string {
	return p.name
}

func (p *socks5ProxyBackend) IsHealth() bool {
	return p.rtt != RTTErrorHappened
}

func (p *socks5ProxyBackend) HealthCheck() error {
	c := &http.Client{Transport: &http.Transport{Dial: p.GetDialer().Dial}}
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

func (p *socks5ProxyBackend) GetPingRTT() time.Duration {
	return p.rtt
}
