package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// tracking latency of requests
var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "loadbalancer_request_duration_seconds",
		Help:    "Duration of HTTP requests through the load balancer",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"path", "method", "status", "backend_id", "strategy"},
)

// tracking the total number of processes requests so far
var TotalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "loadbalancer_total_requests",
		Help: "Total number of requests processed by the load balancer",
	},
	[]string{"path", "method", "status", "backend_id", "strategy"},
)

var BackendStatusGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "loadbalancer_backend_status",
		Help: "Cuurrent health status of backend services(0:unhealthy, 1:healthy)",
	},
	[]string{"backend_host", "backend_id"},
)

var ActiveConnectionsGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "loadbalancer_backend_active_connections",
		Help: "Number of active connections to each backend service",
	},
	[]string{"backend_host", "backend_id"},
)

func InitMetrics() {
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(BackendStatusGauge)
	prometheus.MustRegister(ActiveConnectionsGauge)

	http.Handle("/metrics", promhttp.Handler())
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &responseWriter{ResponseWriter: w}

		ctx := context.WithValue(r.Context(), "responseWriter", wrappedWriter)
		r = r.WithContext(ctx)

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start).Seconds()
		status := wrappedWriter.Status()
		// backend ID and strategy to be used are set by the proxy
		backendID := r.Context().Value("backend_id")
		strategyUsed := r.Context().Value("strategy_used")

		RequestDuration.WithLabelValues(r.URL.Path, r.Method, fmt.Sprintf("%d", status), fmt.Sprintf("%v", backendID), fmt.Sprintf("%v", strategyUsed)).Observe(duration)
		TotalRequests.WithLabelValues(r.URL.Path, r.Method, fmt.Sprintf("%d", status), fmt.Sprintf("%v", backendID), fmt.Sprintf("%v", strategyUsed)).Inc()
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// overriding http.WriteHeader to write a custom header
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}

	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) Status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}