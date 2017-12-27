package commandhelper

import (
	"fmt"
	"os"
)

var Config = NewClientConfig()

type ClientConfig struct {
	AdminLocation string
	Insecure      bool
	OcyDns        string
}

func NewClientConfig() *ClientConfig {
	var adminPort string
	var adminHost string
	var ocyDns string
	if v := os.Getenv("ADMIN_PORT"); v == "" {
		adminPort = "10000"
	} else {
		adminPort = v
	}
	if v := os.Getenv("ADMIN_HOST"); v == "" {
		adminHost = "localhost"
	} else {
		adminHost = v
	}
	if v := os.Getenv("CERT_DNS"); v == "" {
		ocyDns = "ocelot.hq.l11.com"
	} else {
		ocyDns = v
	}
	_, ok := os.LookupEnv("CLIENT_INSECURE")
	if ok {
		fmt.Println("The environment variable CLIENT_INSECURE is set. Using fake certs.")
	}
	return &ClientConfig{
		AdminLocation: adminHost + ":" + adminPort,
		Insecure:      ok,
		OcyDns: ocyDns,
	}
}