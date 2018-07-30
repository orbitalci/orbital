package main

import (
	//"encoding/json"
	"fmt"
	"github.com/namsral/flag"
	ocelog "github.com/shankj3/go-til/log"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/secure_grpc"
	//"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/router/admin"
	//"github.com/shankj3/ocelot/storage"
	"github.com/shankj3/ocelot/version"
	//"io/ioutil"
	"os"
)

func main() {
	//load properties
	var port string
	var gatewayPort string
	var consulHost string
	var consulPort int
	var logLevel string
	var insecure bool

	adminFlags := flag.NewFlagSet("admin", flag.ExitOnError)
	adminFlags.StringVar(&port, "port", "10000", "admin grpc server port")
	adminFlags.StringVar(&gatewayPort, "http-port", "11000", "admin http server port")
	adminFlags.StringVar(&consulHost, "consul-host", "localhost", "consul host")
	adminFlags.IntVar(&consulPort, "consul-port", 8500, "consul port")
	adminFlags.StringVar(&logLevel, "log-level", "debug", "ocelot admin log level")
	adminFlags.BoolVar(&insecure, "insecure", false, "use insecure certs")
	adminFlags.Parse(os.Args[1:])
	version.MaybePrintVersion(adminFlags.Args())

	ocelog.InitializeLog(logLevel)

	serverRunsAt := fmt.Sprintf(":%v", port)
	ocelog.Log().Debug(serverRunsAt)

	configInstance, err := cred.GetInstance(consulHost, consulPort, "")
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("could not talk to consul or vault, bailing")
	}
	var security secure_grpc.SecureGrpc
	if insecure {
		security = secure_grpc.NewFakeSecure()
	} else {
		security = secure_grpc.NewLeSecure()
	}
	grpcServer, listener, store, cancel, err := admin.GetGrpcServer(configInstance, security, serverRunsAt, port, gatewayPort)
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("fatal")
	}
	//if credDumpPath, ok := os.LookupEnv("CRED_DUMP_PATH"); ok {
	//	fmt.Println("dumping everything")
	//	dump_creds(store, configInstance, credDumpPath)
	//}
	defer cancel()
	defer store.Close()
	admin.Start(grpcServer, listener)
}

/*
//fyi this function should be uncommented and used if you want to easily migrate credentials between dbs.
func dump_creds(store storage.OcelotStorage, configInstance cred.CVRemoteConfig, dumpLoc string) {
	allCreds, _ := configInstance.GetAllCreds(store, false)

	allSortedUp := make(map[string][]pb.OcyCredder)
	for _, creddy := range allCreds {
		credtypstr:= creddy.GetSubType().Parent().String()
		_, ok := allSortedUp[credtypstr]
		if !ok {
			allSortedUp[credtypstr] = []pb.OcyCredder{creddy}
		}
		allSortedUp[credtypstr] = append(allSortedUp[credtypstr], creddy)
	}
	allCredsBytes, err := json.Marshal(allSortedUp)
	if err != nil {
		fmt.Println("couldnt dump")
	}
	ioutil.WriteFile(dumpLoc, allCredsBytes, 0644)
}*/