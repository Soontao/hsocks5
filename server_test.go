package hsocks5

import (
	"testing"

	"github.com/armon/go-socks5"
	"github.com/stretchr/testify/assert"
)

func TestNewProxyServer(t *testing.T) {
	socksAddr := "127.0.0.1:50001"
	httpProxyAddr := "127.0.0.1:50002"
	socksServer, err := socks5.New(&socks5.Config{})
	assert.NoError(t, err)

	go func() {
		err = socksServer.ListenAndServe("tcp", socksAddr)
		assert.NoError(t, err)
	}()

	p, err := NewProxyServer(&ProxyServerOption{
		ListenAddr:  httpProxyAddr,
		SocksAddr:   socksAddr,
		ChinaSwitch: true,
	})

	assert.NoError(t, err)
	assert.False(t, p.isWithoutProxy("https://www.google.com"), "'google' should be with proxy")
	assert.True(t, p.isWithoutProxy("192.168.3.1"), "'192.168.3.1' should be without proxy")

}
