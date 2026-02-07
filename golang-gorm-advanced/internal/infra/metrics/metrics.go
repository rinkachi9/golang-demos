package metrics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestCount *prometheus.CounterVec
}

func New() *Metrics {
	m := &Metrics{
		RequestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
	}
	prometheus.MustRegister(m.RequestCount)
	return m
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		status := strconv.Itoa(c.Writer.Status())
		m.RequestCount.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
	}
}

func StartServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	return server
}
