package main

import (
	"github.com/namsral/flag"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/deserialize"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"io/ioutil"
	"net/http"
	"github.com/shankj3/ocelot/util"
	"fmt"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"net"
	"google.golang.org/grpc/credentials"
	"crypto/tls"
	"strings"
	"log"
	"crypto/x509"
	"github.com/philips/grpc-gateway-example/insecure"
)

//TODO: floe integration??? just putting this note here so we remember
//TODO: change this to use my fork of logrus so we can pretty print logs

func main() {
	//load properties
	var port string
	var consulHost string
	var consulPort int
	var logLevel string

	flag.StringVar(&port, "port", "8080", "admin server port")
	flag.StringVar(&consulHost, "consul-host", "localhost", "consul host")
	flag.IntVar(&consulPort, "consul-port", 8500, "consul port")
	flag.StringVar(&logLevel, "log-level", "debug", "ocelot admin log level")
	flag.Parse()

	ocelog.InitializeOcelog(logLevel)

	//TODO: this is my local vault root token, too lazy to set env variable
	configInstance, err := util.GetInstance(consulHost, consulPort, "466e3ddc-f588-d552-6b9a-5f959a9f20a8")

	if err != nil {
		ocelog.Log().Fatal("could not talk to consul or vault, bailing")
	}

	//TODO: figure out if there's a way I can do this without casting back to struct every time
	//initializes our "context" - guideOcelotServer
	guideOcelotServer := NewGuideOcelotServer(configInstance, deserialize.New(), GetValidator(),)

	//check for config on load
	ReadConfig(guideOcelotServer)

	fakeCert := x509.NewCertPool()
	ok := fakeCert.AppendCertsFromPEM([]byte(insecure.Cert))
	//ok := fakeCert.AppendCertsFromPEM([]byte(models.Cert))
	if !ok {
		panic("bad certs")
	}

	pair, err := tls.X509KeyPair([]byte(insecure.Cert), []byte(insecure.Key))
	//pair, err := tls.X509KeyPair([]byte(models.Cert), []byte(models.Key))
	fakeKeyPair := &pair

	//grpc server
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(fakeCert, "localhost:10000"))}

	grpcServer := grpc.NewServer(opts...)
	models.RegisterGuideOcelotServer(grpcServer, guideOcelotServer)
	ctx := context.Background()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: "localhost:10000",
		RootCAs:    fakeCert,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	mux := http.NewServeMux()

	//grpc gateway proxy
	runtime.HTTPError = CustomErrorHandler
	gwmux := runtime.NewServeMux()
	err = models.RegisterGuideOcelotHandlerFromEndpoint(ctx, gwmux, "localhost:10000", dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", 10000))
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:    "localhost:10000",
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*fakeKeyPair},
			NextProtos:   []string{"h2"},
		},
	}

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	//list all repos + 'tracked' repos vs. 'untracked' repos
	//add new repo
	//configure whether or not you want admin to discover new ocelot.yaml files for you
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			ocelog.Log().Debug("serving grpc guuuuurl")
			grpcServer.ServeHTTP(w, r)
		} else {
			ocelog.Log().Debug("serving HTTTTTTTTTTPPPPPP")
			otherHandler.ServeHTTP(w, r)
		}
	})
}

//TODO: damn how to propagate error codes up????
func CustomErrorHandler(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	// see example here: https://github.com/mycodesmells/golang-examples/blob/master/grpc/cmd/server/main.go
	ocenet.JSONApiError(w, runtime.HTTPStatusFromCode(grpc.Code(err)), "", err)
}

//reads config file in current directory if it exists, exits if file is unparseable or doesn't exist
func ReadConfig(gosss models.GuideOcelotServer) {
	gos := gosss.(*guideOcelotServer)

	config := &models.CredWrapper{}
	configFile, err := ioutil.ReadFile(models.ConfigFileName)
	//configFile, err := ioutil.ReadFile("/Users/mariannefeng/go/src/github.com/shankj3/ocelot/admin/" + models.ConfigFileName)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return
	}
	err = gos.Deserializer.YAMLToProto(configFile, config)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return
	}
	for _, configVal := range config.Credentials {
		err := gos.AdminValidator.ValidateConfig(configVal)
		if err != nil {
			ocelog.IncludeErrField(err)
			continue
		}

		err = SetupCredentials(gos, configVal)
		if err != nil {
			ocelog.IncludeErrField(err).Error()
		}
	}
}

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss models.GuideOcelotServer, config *models.Credentials) (error) {
	gos := gosss.(*guideOcelotServer)

	//hehe right now we only have bitbucket
	switch config.Type {
	case "bitbucket":
		bbHandler := handler.Bitbucket{}
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		err := bbHandler.SetMeUp(config, bitbucketClient)

		if err != nil {
			ocelog.Log().Error("could not setup bitbucket client")
			return err
		}

		err = bbHandler.Walk()
		if err != nil {
			return err
		}
	}
	configPath := util.ConfigPath + "/" + config.Type + "/" + config.AcctName
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}