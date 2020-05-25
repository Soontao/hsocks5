package backend

import (
	"sync"
)

// ProxyBackendProvider interface
//
// When create a proxy backend provider, please remember to do health check for all instance firstly
type ProxyBackendProvider interface {
	GetAll() []ProxyBackend
	// GetOne proxy which fastest & health
	GetOne() ProxyBackend
}

// HealthCheckResult enum
type HealthCheckResult int64

const (
	// AllHealth for backends
	AllHealth HealthCheckResult = 1
	// PartialHealth for backends
	PartialHealth = 2
	// FullUnHealth for backends
	FullUnHealth = 3
)

// RunHealthCheck for provider
func RunHealthCheck(pbp ProxyBackendProvider) {

	backends := pbp.GetAll()
	wg := sync.WaitGroup{}
	wg.Add(len(backends))
	for _, b := range backends {
		go func(backend ProxyBackend) {
			backend.HealthCheck() // log error
			wg.Done()
		}(b)
	}
	wg.Wait()

}
