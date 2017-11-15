package nsqpb

import (
	"fmt"
	"os"
)

var (
	// environment variables
	envNsqLookupd = "NSQLOOKUPD_IP"
	envNsqd		  = "NSQD_IP"

	// defaults
	defaultLookupDIP   = "127.0.0.1"
	defaultNsqdIP	   = "127.0.0.1"
	defaultNsqdPort    = "4150"
	defaultLookupDPort = "4161"
)


type NsqConfig struct {
	NsqLookupdIp   string
	NsqdIp		   string
	NsqdPort	   string
	NsqLookupdPort string
}


// LookupDAddress returns `<ip>:<port>` of configured nsqlookupd, the format nsq package takes
func (n *NsqConfig) LookupDAddress() string{
	return fmt.Sprintf("%s:%s", n.NsqLookupdIp, n.NsqLookupdPort)
}

// NsqDAddress returns `<ip>:<port>` of configured nsqd, the format nsq package takes
func (n *NsqConfig) NsqDAddress() string {
	return fmt.Sprintf("%s:%s", n.NsqdIp, n.NsqdPort)
}


// DefaultNsqConf returns new NsqConfig struct with default values.
// Searches environment variables for nsqlookupd ip addr and nsqd ip addr. defaults to 127.0.0.1
// if not found.
func DefaultNsqConf() *NsqConfig {
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