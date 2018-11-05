package du

import (
	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
)

type HCer struct {
	PauseThreshold string
	Path 		   string
}

func (hc *HCer) Healthy() error {
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