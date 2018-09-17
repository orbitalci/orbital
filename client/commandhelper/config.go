package commandhelper

import (
	"fmt"
	models "github.com/shankj3/ocelot/models/pb"
	"os"
	"strconv"
	"strings"
)

var Config *ClientConfig

var (
	// inject with -X github.com/shankj3/ocelot/client/commandhelper.AdminHost=my.ocelotadmin.com -X github.com/shankj3/ocelot/client/commandhelper.AdminPort=443
	AdminHost = "localhost"
	AdminPort = "10000"
)

func init() {
	Config = NewClientConfig()
}

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
		adminPort = AdminPort
	} else {
		adminPort = v
	}
	if v := os.Getenv("ADMIN_HOST"); v == "" {
		adminHost = AdminHost
	} else {
		adminHost = v
	}
	if v := os.Getenv("CERT_DNS"); v == "" {
		ocyDns = "ocyadmin.l11.com"
	} else {
		ocyDns = v
	}
	_, noTLS := os.LookupEnv("NO_USE_TLS")
	//if noTLS {
	//	fmt.Println("not creating https config to make grpc client with")
	//}
	// if the host is an ip address, and NO_USE_TLS was not explicitly set, then set NO_USE_TLS value (noTLS) and print an info screen. if they wan
	if IsIPv4Address(adminHost) && !noTLS {
		if _, yesTLS := os.LookupEnv("USE_TLS"); !yesTLS {
			fmt.Println("Detected an IP address for the admin server host, therefore ocelot will not create https config. \nIf you really wish to try using an https config, then set USE_TLS=true as an environment variable. \nSet NO_USE_TLS=true as an environment variable to suppress this warning in the future. ")
			noTLS = true
		}
	}
	//_, insecure := os.LookupEnv("CLIENT_INSECURE")
	//if insecure {
	//	fmt.Println("The environment variable CLIENT_INSECURE is set. Using fake certs.")
	//}
	client, err := GetClient(adminHost+":"+adminPort, noTLS, ocyDns)
	if err != nil {
		fmt.Println("Could not get client! Error: ", err)
		os.Exit(1)
	}
	_, colorless := os.LookupEnv("NO_COLOR")

	return &ClientConfig{
		AdminLocation: adminHost + ":" + adminPort,
		Client:        client,
		OcyDns:        ocyDns,
		Theme:         Default(colorless),
	}
}

func IsIPv4Address(host string) bool {
	if host == "localhost" {
		return true
	}
	split := strings.Split(host, ".")
	if len(split) < 4 {
		return false
	}
	for _, sub := range split {
		if _, err := strconv.Atoi(sub); err != nil {
			return false
		}
	}
	return true
}
