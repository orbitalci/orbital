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

func startRequest() time.Time {
	activeRequests.Inc()
	return time.Now()
}

func finishRequest(startTime time.Time) {
	requestProcessTime.Observe(time.Since(startTime).Seconds())
	activeRequests.Dec()
}
