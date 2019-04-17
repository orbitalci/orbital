package admin

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestProcessTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ocelot_admin_request_proc_time",
		Help:    "duration of requests send to admin",
		Buckets: prometheus.LinearBuckets(0, 0.25, 15),
	})
	activeRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ocelot_admin_active_requests",
		Help: "number of active requests admin is serving",
	})
)

func StartRequest() time.Time {
	activeRequests.Inc()
	return time.Now()
}

func FinishRequest(startTime time.Time) {
	requestProcessTime.Observe(time.Since(startTime).Seconds())
	activeRequests.Dec()
}

var (
	TriggeredBuilds = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_triggered_builds",
		Help: "builds triggered by a call to admin",
	}, []string{"account", "repository"})
)

func init() {
	prometheus.MustRegister(TriggeredBuilds)
}
