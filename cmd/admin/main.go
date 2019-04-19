package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/namsral/flag"

	"github.com/level11consulting/ocelot/server/grpc/admin"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/server/tls"
	"github.com/level11consulting/ocelot/version"

	ocelog "github.com/shankj3/go-til/log"
)

// FIXME: consistency: consul's host and port, the var name for configInstance
func main() {
	//load properties
	var port string
	var gatewayPort string
	var consulHost string
	var consulPort int
	var logLevel string
	var insecure bool
	var hookhandlerCallbackBase string
	adminFlags := flag.NewFlagSet("admin", flag.ExitOnError)
	adminFlags.StringVar(&port, "port", "10000", "admin grpc server port")
	adminFlags.StringVar(&gatewayPort, "http-port", "11000", "admin http server port")
	adminFlags.StringVar(&consulHost, "consul-host", "localhost", "consul host")
	adminFlags.IntVar(&consulPort, "consul-port", 8500, "consul port")
	adminFlags.StringVar(&logLevel, "log-level", "debug", "ocelot admin log level")
	adminFlags.BoolVar(&insecure, "insecure", false, "use insecure certs")
	adminFlags.StringVar(&hookhandlerCallbackBase, "hookhandler-url-base", "", "base url for registering webhooks for tracking a new repository. this is the url that the hookhandler service is running under, e.g. https://hookhandler.ocelot.io ; do not append with a slash as it will be concatenated and set for each of the supported VCS types")
	adminFlags.Parse(os.Args[1:])
	version.MaybePrintVersion(adminFlags.Args())

	ocelog.InitializeLog(logLevel)

	serverRunsAt := fmt.Sprintf(":%v", port)
	ocelog.Log().Debug(serverRunsAt)

	parsedConsulURL, parsedErr := url.Parse(fmt.Sprintf("consul://%s:%d", consulHost, consulPort))
	if parsedErr != nil {
		ocelog.IncludeErrField(parsedErr).Fatal("failed parsing consul uri, bailing")
	}

	configInstance, err := config.GetInstance(parsedConsulURL, "")
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("could not talk to consul or vault, bailing")
	}
	var security tls.SecureGrpc
	if insecure {
		security = tls.NewFakeSecure()
	} else {
		security = tls.NewLeSecure()
	}
	grpcServer, listener, store, cancel, err := admin.GetGrpcServer(configInstance, security, serverRunsAt, port, gatewayPort, hookhandlerCallbackBase)
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("fatal")
	}

	defer cancel()
	defer store.Close()
	admin.Start(grpcServer, listener)
}
