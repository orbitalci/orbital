package admin

import (
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
		adminPort = "10000"
	} else {
		adminPort = v
	}
	if v := os.Getenv("ADMIN_HOST"); v == "" {
		adminHost = "localhost"
	} else {
		adminHost = v
	}
	return &ClientConfig{
		AdminLocation: adminHost + ":" + adminPort,
	}
}