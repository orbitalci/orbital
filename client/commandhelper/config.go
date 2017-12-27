package commandhelper

import (
	"bitbucket.org/level11consulting/ocelot/admin"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
	"os"
)

var Config = NewClientConfig()

type ClientConfig struct {
	AdminLocation string
	Client        models.GuideOcelotClient
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
	_, insecure := os.LookupEnv("CLIENT_INSECURE")
	if insecure {
		fmt.Println("The environment variable CLIENT_INSECURE is set. Using fake certs.")
	}
	client, err := admin.GetClient(adminHost + ":" + adminPort, insecure, ocyDns)
	if err != nil {
		fmt.Println("Could not get client! Error: ", err)
		os.Exit(1)
	}

	return &ClientConfig{
		AdminLocation: adminHost + ":" + adminPort,
		Client: client,
		OcyDns: ocyDns,
	}
}

