package hsocks5

import (
	"fmt"
	"github.com/Soontao/hsocks5/backend"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyServer(t *testing.T) {

	socksAddr := "127.0.0.1:50001"
	httpProxyAddr := "127.0.0.1:50002"
	httpProxyURL, err := url.Parse("http://127.0.0.1:50002")
	assert.NoError(t, err)
	// mock socks server
	socksServer, err := socks5.New(&socks5.Config{})
	assert.NoError(t, err)

	l, err := net.Listen("tcp", socksAddr)
	assert.NoError(t, err)

	go func() {
		e := socksServer.Serve(l)
		assert.NoError(t, e)
	}()

	p, err := NewProxyServer(&ProxyServerOption{
		ListenAddr:  httpProxyAddr,
		SocksAddr:   socksAddr,
		ChinaSwitch: true,
	})

	go func() {
		e := p.Start()
		assert.NoError(t, e)
	}()

	time.Sleep(100 * time.Microsecond) // wait some micro-seconds, make sever started

	assert.NoError(t, err)
	assert.False(t, p.isWithoutProxy("https://www.google.com"), "'google' should be with proxy")
	assert.False(t, p.isWithoutProxy("https://www.google.com"), "'google' should be cached")
	assert.False(t, p.isWithoutProxy("172.217.160.78"), "'google' ip should be with proxy")
	assert.True(t, p.isWithoutProxy("192.168.3.1"), "'192.168.3.1' should be without proxy")

	req := httptest.NewRequest("GET", "https://github.com", nil)
	w := httptest.NewRecorder()
	p.handleProxyRequest(w, req)
	resp := w.Result()
	assert.Equal(t, 200, resp.StatusCode)

	req = httptest.NewRequest("Get", "/hsocks5/__/metric", nil)
	w = httptest.NewRecorder()
	p.handleRequest(w, req)
	resp = w.Result()
	assert.Equal(t, 200, resp.StatusCode)

	req = httptest.NewRequest("Get", "/whatever", nil)
	w = httptest.NewRecorder()
	p.handleRequest(w, req)
	resp = w.Result()
	assert.Equal(t, 404, resp.StatusCode)

	c := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(httpProxyURL)}}
	req, err = http.NewRequest("GET", "https://github.com", nil)
	assert.NoError(t, err)
	resp, err = c.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	req, err = http.NewRequest(
		"GET",
		fmt.Sprintf("http://%v/hsocks5/__/metric", httpProxyAddr),
		nil,
	)
	assert.NoError(t, err)
	resp, err = c.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	assert2 := assert.New(t)
	assert2.NoError(err)

	b1, err := backend.NewSocks5ProxyBackend(&backend.Socks5ProxyBackendOption{
		HealthEndpoint: "https://github.com",
		Addr:           socksAddr,
		Name:           "mock-socks-server",
	})

	assert2.NoError(err)

	backendProvider, err := backend.NewFastProxyBackendProvider(&backend.FastProxyBackendOption{
		Backends: []backend.ProxyBackend{b1},
	})

	assert2.NoError(err)

	p2, err := NewProxyServer(&ProxyServerOption{
		ListenAddr:  httpProxyAddr,
		SocksAddr:   socksAddr,
		ChinaSwitch: true,
		Provider:    backendProvider,
	})

	req = httptest.NewRequest("GET", "https://github.com", nil)
	w = httptest.NewRecorder()
	p2.handleProxyRequest(w, req)
	resp = w.Result()
	assert.Equal(t, 200, resp.StatusCode)

}
