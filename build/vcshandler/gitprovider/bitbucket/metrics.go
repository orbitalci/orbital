package bitbucket

import "github.com/prometheus/client_golang/prometheus"

var (
	failedBBRemoteCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_bitbucket_failed_calls",
		Help: "All failed calls to bitbucket",
	}, []string{"method"},)
	tooManyChangesets = prometheus.NewCounter(prometheus.CounterOpts{
Name: "ocelot_bitbucket_too_many_changesets",
Help: "Incidents where there is more than one changeset in a push hook",
},)
	noChangesets = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ocelot_bitbucket_no_changesets",
		Help: "Incidents where there are no changesets in the push hook",
	},)
	noPreviousHead = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ocelot_bitbucket_no_previous_head",
		Help: "Translation occurence where there is no previous head commit in the push json",
	},)
	unsupportedPush = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_bitbucket_unsupported_push",
		Help: "Unsupported type of push",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(failedBBRemoteCalls)
}
