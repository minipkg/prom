package prometheus_utils

import "github.com/prometheus/client_golang/prometheus"

const (
	labelCalled  = "called"
	labelFailed  = "failed"
	labelSucceed = "succeed"
)

// workerMetrics stores any worker metrics: number of worker being called and failed
// succeed can be derived as: called - failed or counted with its own label
type workerMetrics struct {
	*prometheus.CounterVec
}

// NewWorkerMetrics creates a new workerMetrics with the given app and worker name
// The only label is 'status' i.e. called/failed
func NewWorkerMetrics(appName, workerName string) *workerMetrics {
	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "worker_metric",
			Help:        "Number of times worker has the specified in label status",
			ConstLabels: prometheus.Labels{"app": appName, "worker": workerName},
		},
		[]string{"status"},
	)

	prometheus.MustRegister(c)
	return &workerMetrics{c}
}

// Called increments the called counter by 1
func (m *workerMetrics) Called() {
	m.WithLabelValues(labelCalled).Inc()
}

// Failed increments the failed counter by 1
func (m *workerMetrics) Failed() {
	m.WithLabelValues(labelFailed).Inc()
}

// Succeed increments the succeed counter by 1
func (m *workerMetrics) Succeed() {
	m.WithLabelValues(labelSucceed).Inc()
}
