package main

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/admin"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/secure_grpc"
	"fmt"
	"github.com/namsral/flag"
)

func main() {
	//load properties
	var port string
	var consulHost string
	var consulPort int
	var logLevel string
	var insecure bool

	flag.StringVar(&port, "port", "10000", "admin server port")
	flag.StringVar(&consulHost, "consul-host", "localhost", "consul host")
	flag.IntVar(&consulPort, "consul-port", 8500, "consul port")
	flag.StringVar(&logLevel, "log-level", "debug", "ocelot admin log level")
	flag.BoolVar(&insecure, "insecure", false, "use insecure certs")
	flag.Parse()

	ocelog.InitializeLog(logLevel)

	serverRunsAt := fmt.Sprintf("localhost:%v", port)
	ocelog.Log().Debug(serverRunsAt)

	configInstance, err := cred.GetInstance(consulHost, consulPort, "")

	if err != nil {
		ocelog.Log().Fatal("could not talk to consul or vault, bailing")
	}
	var security secure_grpc.SecureGrpc
	if insecure {
		security = secure_grpc.NewFakeSecure()
	} else {
		security = secure_grpc.NewLeSecure()
	}
	admin.Start(configInstance, security, serverRunsAt, port)

}
