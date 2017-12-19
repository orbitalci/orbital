package admin

import (
	"github.com/namsral/flag"
	"os"
)

var Config = NewClientConfig()

type ClientConfig struct {
	AdminLocation string
}

func NewClientConfig() *ClientConfig {
	var adminPort string
	var adminHost string
	if v := os.Getenv("ADMIN_PORT"); v == "" {
		flag.StringVar(&adminHost, "admin-port", "10000", "ip or fqdn of host for the admin server")
	} else {
		adminPort = v
	}
	if v := os.Getenv("ADMIN_HOST"); v == "" {
		flag.StringVar(&adminPort, "admin-host", "localhost", "port on which admin server is running")
	} else {
		adminHost = v
	}
	flag.Parse()
	return &ClientConfig{
		AdminLocation: adminHost + ":" + adminPort,
	}
}