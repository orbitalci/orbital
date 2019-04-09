package common

import (
	"strings"

	"github.com/level11consulting/ocelot/models/pb"
)

func NCErr(msg string) *NoCreds {
	return &NoCreds{msg: msg}
}

type NoCreds struct {
	msg string
}

func (n *NoCreds) Error() string {
	return n.msg
}

// BuildScriptsContainString will check all stages' script lines for the existence of the specified desiredString
func BuildScriptsContainString(wc *pb.BuildConfig, desiredString string) bool {
	for _, stage := range wc.Stages {
		for _, script := range stage.Script {
			if strings.Contains(script, desiredString) {
				return true
			}
		}
	}
	return false
}
