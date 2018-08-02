package common

import (
	"encoding/base64"
	"strings"

	"github.com/shankj3/ocelot/models/pb"
)

func NCErr(msg string) *NoCreds {
	return &NoCreds{msg:msg}
}

type NoCreds struct {
	msg string
}

func (n *NoCreds) Error() string {
	return n.msg
}

func BitzToBase64(bits []byte) string {
	return base64.StdEncoding.EncodeToString(bits)
}

func StrToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64ToBitz(b64string string) ([]byte, error) {
	 return base64.StdEncoding.DecodeString(b64string)
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