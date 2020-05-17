package hsocks5

import "github.com/prometheus/client_golang/prometheus"

// ProxyServerMetrics class
type ProxyServerMetrics struct {
	connTotal          *prometheus.CounterVec
	requestStatusTotal *prometheus.CounterVec
	cacheHitTotal      *prometheus.CounterVec
	routineResultTotal *prometheus.CounterVec
}

// NewProxyServerMetrics constructor
func NewProxyServerMetrics() (rt *ProxyServerMetrics) {
	rt = &ProxyServerMetrics{}

	rt.connTotal = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "hsocks5_conn_total", Help: "Total Connections"}, []string{"type"})

	rt.requestStatusTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "hsocks5_request_status_total", Help: "Request Status"},
		[]string{"hostname", "status"},
	)

	rt.cacheHitTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "hsocks5_routine_cache_hit_total", Help: "Proxy Routing Cache Hit Counter"},
		[]string{"type"},
	)

	rt.routineResultTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "hsocks5_routine_result_total", Help: "Proxy Routing Result"},
		[]string{"hostname", "result", "reason"},
	)

	prometheus.MustRegister(rt.connTotal)
	prometheus.MustRegister(rt.requestStatusTotal)
	prometheus.MustRegister(rt.cacheHitTotal)
	prometheus.MustRegister(rt.routineResultTotal)

	return
}
