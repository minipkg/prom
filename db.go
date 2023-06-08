package prometheus_utils

import (
	"database/sql"
	"github.com/prometheus/client_golang/prometheus"
)

type DbMetrics interface {
	ReadStatsFromDB(s *sql.DB)
}

type dbMetrics struct {
	NbMaxConns      prometheus.Gauge
	NbOpenConns     prometheus.Gauge
	NbUsedConns     prometheus.Gauge
	WaitCount       prometheus.Gauge
	WaitDurationSec prometheus.Summary
}

var _ DbMetrics = (*dbMetrics)(nil)

func NewDbMetrics(namespace, subsystem, service, host, dbName string) *dbMetrics {
	return &dbMetrics{
		NbMaxConns:      newGauge(namespace, subsystem, service, host, dbName, "nb_max_conns", "Maximum number of open connections to the database."),
		NbOpenConns:     newGauge(namespace, subsystem, service, host, dbName, "nb_open_conns", "The number of established connections both in use and idle."),
		NbUsedConns:     newGauge(namespace, subsystem, service, host, dbName, "nb_used_conns", "The number of connections currently in use."),
		WaitCount:       newGauge(namespace, subsystem, service, host, dbName, "wait_count", "The total number of connections waited for."),
		WaitDurationSec: newSummary(namespace, subsystem, service, host, dbName, "wait_duration_sec", "The total time blocked waiting for a new connection (in seconds)."),
	}
}

func newGauge(namespace, subsystem, service, host, dbName, name, help string) prometheus.Gauge {
	g := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
			ConstLabels: prometheus.Labels{
				"namespace": namespace,
				"subsystem": subsystem,
				"service":   service,
				"host":      host,
				"db":        dbName,
			},
		})
	prometheus.MustRegister(g)
	return g
}

func newSummary(namespace, subsystem, service, host, dbName, name, help string) prometheus.Summary {
	s := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: name,
			Help: help,
			ConstLabels: prometheus.Labels{
				"namespace": namespace,
				"subsystem": subsystem,
				"service":   service,
				"host":      host,
				"db":        dbName,
			},
		})
	prometheus.MustRegister(s)
	return s
}

// Метод безопасно вызывать на закрытой БД
func (d *dbMetrics) ReadStatsFromDB(s *sql.DB) {
	stats := s.Stats()

	d.NbMaxConns.Set(float64(stats.MaxOpenConnections))
	d.NbOpenConns.Set(float64(stats.OpenConnections))
	d.NbUsedConns.Set(float64(stats.InUse))
	d.WaitCount.Set(float64(stats.WaitCount))
	d.WaitDurationSec.Observe(stats.WaitDuration.Seconds())

}
