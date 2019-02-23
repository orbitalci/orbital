package credentials

import (
	"github.com/shankj3/go-til/consul"
)

// getConsulAddr will set the Vault address in this order:
// Passing in Consul through command line options takes priority
// If not passed in, the CONSUL_HTTP_ADDR environment variable is next
// If not defined, assume http://localhost:8500
//func getConsulAddr() error {
//}

// GetConsul returns the local consul client handler
func (rc *RemoteConfig) GetConsul() consul.Consuletty {
	return rc.Consul
}

// SetConsul sets the local consul client handler
func (rc *RemoteConfig) SetConsul(consl consul.Consuletty) {
	rc.Consul = consl
}
