package main

import (
	"github.com/shankj3/ocelot/util/ocelog"
	"fmt"
	"github.com/namsral/flag"
	"github.com/shankj3/ocelot/util"
	"github.com/shankj3/ocelot/admin"
)

func main() {
	//load properties
	var port string
	var consulHost string
	var consulPort int
	var logLevel string

	flag.StringVar(&port, "port", "10000", "admin server port")
	flag.StringVar(&consulHost, "consul-host", "localhost", "consul host")
	flag.IntVar(&consulPort, "consul-port", 8500, "consul port")
	flag.StringVar(&logLevel, "log-level", "debug", "ocelot admin log level")
	flag.Parse()

	ocelog.InitializeOcelog(logLevel)

	serverRunsAt := fmt.Sprintf("localhost:%v", port)
	ocelog.Log().Debug(serverRunsAt)

	//TODO: this is my local vault root token, too lazy to set env variable
	configInstance, err := util.GetInstance(consulHost, consulPort, "")

	if err != nil {
		ocelog.Log().Fatal("could not talk to consul or vault, bailing")
	}

	admin.Start(configInstance, serverRunsAt, port)

}
