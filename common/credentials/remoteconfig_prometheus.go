package credentials

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	failedCredRetrieval = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ocelot_failed_cred",
			Help: "number of times ocelot is unable to interact with cred store",
		},
		// interaction_type: create | read | update | delete
		// cred_type: subcredType.String()
		// is_secret: bool , whether or not interacting with secret store or insecure cred store
		[]string{"cred_type", "interaction_type", "is_secret"},
	)
)

func init() {
	prometheus.MustRegister(failedCredRetrieval)
}
