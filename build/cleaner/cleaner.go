package cleaner

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/level11consulting/ocelot/models"
)

//this interface handles build cleanup
type Cleaner interface {

	//Cleanup performs build cleanup functions. If an optional logout channel is passed, logs will be sent over the channel
	Cleanup(ctx context.Context, id string, logout chan []byte) error
}

//returns a new cleaner interface
func GetNewCleaner(werkerType models.WerkType, facts *models.SSHFacts) Cleaner {
	switch werkerType {
	case models.Docker:
		return &DockerCleaner{}
	case models.Kubernetes:
		return &K8Cleaner{}
	case models.SSH:
		return &SSHCleaner{SSHFacts: facts}
	case models.Exec:
		return NewExecCleaner()
	default:
		return &DockerCleaner{}
	}
	return nil
}

var (
	failedCleaning = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_build_clean_failed",
		Help: "post build clean failures",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(failedCleaning)
}
