package main

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/admin"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/secure_grpc"
	"fmt"
	"github.com/namsral/flag"
	"os"
)

func main() {
	//load properties
	var port string
	var consulHost string
	var consulPort int
	var logLevel string
	var insecure bool

	adminFlags := flag.NewFlagSet("admin", flag.ExitOnError)
	adminFlags.StringVar(&port, "port", "10000", "admin server port")
	adminFlags.StringVar(&consulHost, "consul-host", "localhost", "consul host")
	adminFlags.IntVar(&consulPort, "consul-port", 8500, "consul port")
	adminFlags.StringVar(&logLevel, "log-level", "debug", "ocelot admin log level")
	adminFlags.BoolVar(&insecure, "insecure", false, "use insecure certs")
	adminFlags.Parse(os.Args[1:])

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
