package bitbucket

import "github.com/prometheus/client_golang/prometheus"

var (
	failedBBRemoteCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bitbucket_failed_calls",
		Help: "All failed calls to bitbucket",
	}, []string{"method"},
	)
)

func init() {
	prometheus.MustRegister(failedBBRemoteCalls)
}

