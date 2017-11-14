package nsqpb

import (
	"fmt"
	"os"
)

type NsqConfig struct {
	NsqLookupdIp   string
	NsqdIp		   string
	NsqdPort	   string
	NsqLookupdPort string
}

func (n *NsqConfig) LookupDAddress() string{
	return fmt.Sprintf("%s:%s", n.NsqLookupdIp, n.NsqLookupdPort)
}

func (n *NsqConfig) NsqDAddress() string {
	return fmt.Sprintf("%s:%s", n.NsqdIp, n.NsqdPort)
}

var (
	envNsqLookupd = "NSQLOOKUPD_IP"
	envNsqd		  = "NSQD_IP"

	// defaults, yooooo
	defaultLookupDIP   = "127.0.0.1"
	defaultNsqdIP	   = "127.0.0.1"
	defaultNsqdPort    = "4150"
	defaultLookupDPort = "4161"
)

func NewNsqConf() *NsqConfig {
	var nsqlookupd, nsqd string
	// NSQLOOKUPD_IP may have to be looked up more than nsqd_ip, since nsqlookupd
	// likely isn't running everywhere.
	if nsqlookupd = os.Getenv(envNsqLookupd); nsqlookupd == "" {
		nsqlookupd = defaultLookupDIP
	}
	if nsqd = os.Getenv(envNsqd); nsqd == "" {
		nsqd = defaultNsqdIP
	}
	return &NsqConfig{
		NsqLookupdIp:   nsqlookupd,
		NsqdIp: 	    nsqd,
		NsqdPort:       defaultNsqdPort,  // can change these to be configurable later
		NsqLookupdPort: defaultLookupDPort,  // can change these to be configurable later.
	}

}