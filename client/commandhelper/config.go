package commandhelper

import (
	models "bitbucket.org/level11consulting/ocelot/models/pb"
	"fmt"
	"os"
)

var Config = NewClientConfig()

type ClientConfig struct {
	AdminLocation string
	Client        models.GuideOcelotClient
	OcyDns        string
	Theme         *ColorDefs
}

func NewClientConfig() *ClientConfig {
	// todo: add these as actual flagsets, then merge them with the command-specific ones
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
		ocyDns = "ocyadmin.l11.com"
	} else {
		ocyDns = v
	}
	_, insecure := os.LookupEnv("CLIENT_INSECURE")
	if insecure {
		fmt.Println("The environment variable CLIENT_INSECURE is set. Using fake certs.")
	}
	client, err := GetClient(adminHost + ":" + adminPort, insecure, ocyDns)
	if err != nil {
		fmt.Println("Could not get client! Error: ", err)
		os.Exit(1)
	}
	_, colorless := os.LookupEnv("NO_COLOR")

	return &ClientConfig{
		AdminLocation: adminHost + ":" + adminPort,
		Client: client,
		OcyDns: ocyDns,
		Theme: Default(colorless),
	}
}

