package github

import "github.com/prometheus/client_golang/prometheus"

var (
	failedGHRemoteCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_github_failed_calls",
		Help: "All failed calls to github",
	}, []string{"method"},
	)
	noChangesets = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ocelot_github_no_changesets",
		Help: "Incidents where there are no changesets in the push hook",
	},)
	noPreviousHead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ocelot_github_no_previous_head",
		Help: "Translation occurence where there is no previous head commit in the push json",
	},)
	unsupportedPush = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_github_unsupported_push",
		Help: "Unsupported type of push",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(failedGHRemoteCalls)
}
