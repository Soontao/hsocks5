package backend

// ProxyBackendProvider interface
//
// When create a proxy backend provider, please remember to do health check for all instance firstly
type ProxyBackendProvider interface {
	GetAll() []ProxyBackend
	// GetOne proxy which fastest & health
	GetOne() ProxyBackend
}
