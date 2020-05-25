package backend

import (
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"github.com/stretchr/testify/assert"
)

func TestNewSocks5ProxyBackend(t *testing.T) {
	it := assert.New(t)

	// mock socks server
	socksAddr := "127.0.0.1:50001"
	socksServer, err := socks5.New(&socks5.Config{})
	assert.NoError(t, err)

	go func() {
		err := socksServer.ListenAndServe("tcp", socksAddr)
		it.NoError(err)
	}()

	time.Sleep(10 * time.Millisecond)

	backend, err := NewSocks5ProxyBackend(&Socks5ProxyBackendOption{
		Addr:           socksAddr,
		Name:           "Test001",
		HealthEndpoint: "https://github.com",
	})

	it.NoError(err)

	it.NoError(backend.HealthCheck())

	it.Equal("Test001", backend.GetName())
	it.NotEqual(RTTErrorNotCheck, backend.GetPingRTT())
	it.NotEqual(RTTErrorHappened, backend.GetPingRTT())

}
