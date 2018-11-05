package du

import (
	"testing"
)

func TestHCer_Healthy(t *testing.T) {
	hc := &HealthChecker{
		Path: "/var/lib",
		PauseThreshold: "10GB",
	}
	err := hc.Healthy()
	if err != nil {
		t.Error(err)
	}
}
