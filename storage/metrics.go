package storage

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	activeRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ocelot_db_active_requests",
		Help: "number of current db requests",
	})
	dbDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ocelot_db_transaction_duration",
		Help:    "database execution times",
		Buckets: prometheus.LinearBuckets(0, 0.25, 15),
		// table: build_summary, etc
		// interaction_type: create | read | update | delete
	}, []string{"table", "interaction_type"})
	DatabaseFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_db_sqllib_error",
		Help: "sql library error count",
	}, []string{"error_type"})
)


func init() {
	prometheus.MustRegister(activeRequests, dbDuration, DatabaseFailed)
	// seed data
	DatabaseFailed.WithLabelValues("ErrConnDone").Add(0)
	DatabaseFailed.WithLabelValues("ErrBadCon").Add(0)
	DatabaseFailed.WithLabelValues("ErrTxDone").Add(0)
}

func StartTransaction() time.Time {
	activeRequests.Inc()
	return time.Now()
}

func FinishTransaction(start time.Time, table, crud string) {
	activeRequests.Dec()
	dbDuration.WithLabelValues(table, crud).Observe(time.Since(start).Seconds())
}
