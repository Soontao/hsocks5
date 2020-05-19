package hsocks5

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/armon/go-socks5"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyServer(t *testing.T) {
	socksAddr := "127.0.0.1:50001"
	httpProxyAddr := "127.0.0.1:50002"
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
}
