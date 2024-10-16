package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path"},
	)
)

func SetupMetrics(router *gin.Engine) {
	prometheus.MustRegister(reqTotal)

	router.Use(func(c *gin.Context) {
		reqTotal.WithLabelValues(c.FullPath()).Inc()
		c.Next()
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}