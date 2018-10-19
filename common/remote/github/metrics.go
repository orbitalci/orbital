package github

import "github.com/prometheus/client_golang/prometheus"

var (
	failedGHRemoteCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_github_failed_calls",
		Help: "All failed calls to github",
	}, []string{"method"},
	)
)

func init() {
	prometheus.MustRegister(failedGHRemoteCalls)
}
