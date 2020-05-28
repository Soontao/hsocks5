package hsocks5

import "github.com/prometheus/client_golang/prometheus"

// ProxyServerMetrics class
type ProxyServerMetrics struct {
	connTotal          *prometheus.CounterVec
	requestStatusTotal *prometheus.CounterVec
	cacheHitTotal      *prometheus.CounterVec
	routineResultTotal *prometheus.CounterVec
	errorTotal         *prometheus.CounterVec
	trafficSizeTotal   *prometheus.CounterVec
}

var metricSingleInstace *ProxyServerMetrics

// NewProxyServerMetrics constructor
func NewProxyServerMetrics() (rt *ProxyServerMetrics) {
	if metricSingleInstace != nil {
		return metricSingleInstace
	}

	rt = &ProxyServerMetrics{}

	rt.connTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "hsocks5_conn_total", Help: "Total Connections"},
		[]string{"type"},
	)

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

	rt.errorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "hsocks5_error_total", Help: "HSOCKS all errors"},
		[]string{"category", "context"},
	)

	rt.trafficSizeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "hsocks5_traffic_total", Help: "HSOCKS traffic size (byte)"},
		[]string{"type"},
	)

	prometheus.MustRegister(
		rt.connTotal,
		rt.requestStatusTotal,
		rt.cacheHitTotal,
		rt.routineResultTotal,
		rt.errorTotal,
		rt.trafficSizeTotal,
	)

	metricSingleInstace = rt

	return
}

// AddErrorMetric count
func (m *ProxyServerMetrics) AddErrorMetric(category, context string) {
	m.errorTotal.WithLabelValues(category, context).Inc()
}
