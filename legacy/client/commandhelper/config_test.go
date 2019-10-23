package commandhelper

import (
	"fmt"
	"testing"
)

var ipv4s = []struct {
	host string
	isIp bool
}{
	{"10.1.72.229", true},
	{"10.1..229", false},
	{"ocelot-admin-grpc.metaverse.test.com", false},
	{"0.0.0.0", true},
}

func TestIsIPv4Address(t *testing.T) {
	for ind, tt := range ipv4s {
		t.Run(fmt.Sprintf("%d", ind), func(t *testing.T) {
			isIp := IsIPv4Address(tt.host)
			if isIp != tt.isIp {
				t.Errorf("expected isIp to be %#v for host %s, is %#v", tt.isIp, tt.host, isIp)
			}
		})
	}
}
