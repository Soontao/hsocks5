package backend

import (
	"sort"
	"sync"
	"time"
)

// FastProxyBackendOption type
type FastProxyBackendOption struct {
	HealthCheckInterval time.Duration
	Backends            []ProxyBackend
}

// DefaultFastProxyHealthCheckInterval is 60 seconds
const DefaultFastProxyHealthCheckInterval = time.Second * 60

// NewFastProxyBackendProvider instance
//
// this provider will always provide the fastest proxy backend instance
func NewFastProxyBackendProvider(option *FastProxyBackendOption) (ProxyBackendProvider, error) {

	// default interval
	if option.HealthCheckInterval == 0 {
		option.HealthCheckInterval = DefaultFastProxyHealthCheckInterval
	}

	// not provide backend
	if len(option.Backends) == 0 {
		return nil, ErrBackendNotProvided
	}

	rt := &fastProxyBackendProvider{
		backends:            option.Backends,
		rw:                  &sync.RWMutex{},
		jobLock:             &sync.Mutex{},
		healthCheckInterval: option.HealthCheckInterval,
		jobStarted:          false,
	}

	rt.healthCheckAll() // must do health check all firstly
	rt.sort()

	return rt, nil
}

type fastProxyBackendProvider struct {
	backends            []ProxyBackend
	rw                  *sync.RWMutex
	jobLock             *sync.Mutex
	healthCheckInterval time.Duration
	jobStarted          bool
}

// startJob
func (pbp *fastProxyBackendProvider) startJob() {

	if pbp.jobStarted { // job has been started
		return
	}

	pbp.jobStarted = true

	go func() {
		for range time.Tick(pbp.healthCheckInterval) {
			func() {

				pbp.jobLock.Lock()
				defer pbp.jobLock.Unlock()

				pbp.healthCheckAll() // do health check for all proxy backend
				pbp.sort()           // sort by PingRTT

			}()
		}
	}()

}

// healthCheckAll, do health check & sort backends by PingRTT asc
func (pbp *fastProxyBackendProvider) healthCheckAll() {

	wg := sync.WaitGroup{}
	wg.Add(len(pbp.backends))
	for _, b := range pbp.backends {
		go func(backend ProxyBackend) {
			backend.HealthCheck() // log error
			wg.Done()
		}(b)
	}
	wg.Wait()

}

// sort all providers
func (pbp *fastProxyBackendProvider) sort() {

	pbp.rw.Lock()
	defer pbp.rw.Unlock()
	sort.Sort(byPingRTT(pbp.backends))

}

func (pbp *fastProxyBackendProvider) GetAll() []ProxyBackend {
	pbp.rw.RLock()
	defer pbp.rw.RUnlock()
	return pbp.backends
}

func (pbp *fastProxyBackendProvider) GetOne() ProxyBackend {
	pbp.rw.RLock()
	defer pbp.rw.RUnlock()
	return pbp.backends[0]
}

// ByDelay sort
type byPingRTT []ProxyBackend

func (a byPingRTT) Len() int           { return len(a) }
func (a byPingRTT) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPingRTT) Less(i, j int) bool { return a[i].GetPingRTT() < a[j].GetPingRTT() }
