package du

import (
	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
)

type HealthChecker struct {
	PauseThreshold string
	Path 		   string
}

func (hc *HealthChecker) Healthy() error {
	if hc.PauseThreshold == "" || hc.Path == "" {
		return nil
	}
	thresholdBytes, err := bytefmt.ToBytes(hc.PauseThreshold)
	if err != nil {
		return err
	}
	_, free, err := Space(hc.Path)
	if err != nil {

		return err
	}
	if thresholdBytes > free {
		return errors.Errorf("Free space (%s) is less than threshold (%s). This is unhealthy.", bytefmt.ByteSize(free), hc.PauseThreshold)
	}
	return nil
}