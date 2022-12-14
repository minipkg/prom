package prometheus_utils

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
)

const (
	AuthClientKey = "http.client"
)

// httpServerMetrics is a struct that allows to write metrics of count and latency of http requests
type httpServerMetric struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

func NewHttpServerMetrics(appName string) *httpServerMetric {
	reqsCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "reqs_count",
			Help:        "How many HTTP requests processed",
			ConstLabels: prometheus.Labels{"app": appName},
		},
		[]string{"method", "status", "path", "client"},
	)

	latencyCollector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "reqs_latency",
		Help:        "How long it took to process the request",
		ConstLabels: prometheus.Labels{"app": appName},
		Buckets:     []float64{5, 10, 20, 30, 50, 70, 100, 150, 200, 300, 500, 1000},
	},
		[]string{"method", "status", "path", "client"},
	)

	prometheus.MustRegister(reqsCollector, latencyCollector)

	return &httpServerMetric{
		reqs:    reqsCollector,
		latency: latencyCollector,
	}
}

// Inc increases requests counter by one.
//  method, code, path and client are label values for "method", "status", "path" and "client" fields
func (m *httpServerMetric) Inc(method, code, path, client string) {
	m.reqs.WithLabelValues(method, code, path, client).Inc()
}

// WriteTiming writes time elapsed since the startTime.
// method, code, path and client are label values for "method", "status", "path" and "client" fields
func (m *httpServerMetric) WriteTiming(startTime time.Time, method, code, path, client string) {
	m.latency.WithLabelValues(method, code, path, client).Observe(timeFromStart(startTime))
}

// Handler with metrics for "github.com/fasthttp/router"
func GetFasthttpHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
}

// Middleware with metrics for "github.com/fasthttp/router"
func (m *httpServerMetric) FasthttpRouterMetricsMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		now := time.Now()

		next(ctx)

		client := ""
		if s, ok := ctx.UserValue(AuthClientKey).(string); ok {
			client = s
		}

		status := strconv.Itoa(ctx.Response.StatusCode())
		path := string(ctx.Path())
		method := string(ctx.Method())

		m.Inc(method, status, path, client)
		m.WriteTiming(now, method, status, path, client)
	}
}

// Handler with metrics for "github.com/qiangxue/fasthttp-routing"
func GetFasthttpRoutingHandler() routing.Handler {
	return func(rctx *routing.Context) error {
		fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(rctx.RequestCtx)
		return nil
	}
}

// Middleware with metrics for "github.com/qiangxue/fasthttp-routing"
func (m *httpServerMetric) FasthttpRoutingMetricsMiddleware(rctx *routing.Context) {
	now := time.Now()

	rctx.Next()

	client := ""
	if s, ok := rctx.UserValue(AuthClientKey).(string); ok {
		client = s
	}

	status := strconv.Itoa(rctx.Response.StatusCode())
	path := string(rctx.Path())
	method := string(rctx.Method())

	m.Inc(method, status, path, client)
	m.WriteTiming(now, method, status, path, client)
}
