package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/admin/handler"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/secure_grpc"
	"crypto/tls"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

//TODO: floe integration??? just putting this note here so we remember
//TODO: change this to use my fork of logrus so we can pretty print logs?

//Start will kick off our grpc server so it's ready to receive requests over both grpc and http
func Start(configInstance *cred.RemoteConfig, secure secure_grpc.SecureGrpc, serverRunsAt string, port string) {
	//initializes our "context" - guideOcelotServer
	guideOcelotServer := NewGuideOcelotServer(configInstance, deserialize.New(), GetValidator())

	//check for config on load
	ReadConfig(guideOcelotServer)

	//grpc server
	opts := []grpc.ServerOption{
		grpc.Creds(secure.GetNewClientTLS(serverRunsAt))}

	grpcServer := grpc.NewServer(opts...)
	models.RegisterGuideOcelotServer(grpcServer, guideOcelotServer)

	ctx := context.Background()



	dopts := []grpc.DialOption{grpc.WithTransportCredentials(secure.GetNewTLS(serverRunsAt))}
	mux := http.NewServeMux()

	runtime.HTTPError = CustomErrorHandler
	gwmux := runtime.NewServeMux()
	err := models.RegisterGuideOcelotHandlerFromEndpoint(ctx, gwmux, serverRunsAt, dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:    serverRunsAt,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*secure.GetKeyPair()},
			NextProtos:   []string{"h2"},
		},
	}

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Log().Fatal("ListenAndServe: ", err)
	}
}

// see full example at: https://github.com/philips/grpc-gateway-example

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			log.Log().Debug("serving grpc guuuuurl")
			grpcServer.ServeHTTP(w, r)
		} else {
			log.Log().Debug("serving HTTTTTTTTTTPPPPPP")
			otherHandler.ServeHTTP(w, r)
		}
	})
}

//TODO: how to propagate error codes up????
//TODO: cast this back to MY error type and set status
func CustomErrorHandler(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	// see example here: https://github.com/mycodesmells/golang-examples/blob/master/grpc/cmd/server/main.go
	ocenet.JSONApiError(w, runtime.HTTPStatusFromCode(grpc.Code(err)), "", err)
}

//reads config file in current directory if it exists, exits if file is unparseable or doesn't exist
func ReadConfig(gosss models.GuideOcelotServer) {
	gos := gosss.(*guideOcelotServer)

	config := &models.CredWrapper{}
	configFile, err := ioutil.ReadFile("/home/mariannefeng/go/src/bitbucket.org/level11consulting/ocelot/admin/" + models.ConfigFileName)
	//configFile, err := ioutil.ReadFile("/Users/mariannefeng/go/src/bitbucket.org/level11consulting/ocelot/admin/" + models.ConfigFileName)
	if err != nil {
		log.IncludeErrField(err).Error()
		return
	}
	err = gos.Deserializer.YAMLToProto(configFile, config)
	if err != nil {
		log.IncludeErrField(err).Error()
		return
	}
	for _, configVal := range config.Credentials {
		err := gos.AdminValidator.ValidateConfig(configVal)
		if err != nil {
			log.IncludeErrField(err)
			continue
		}

		err = SetupCredentials(gos, configVal)
		if err != nil {
			log.IncludeErrField(err).Error()
		}
	}
}

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss models.GuideOcelotServer, config *models.Credentials) error {
	gos := gosss.(*guideOcelotServer)

	//hehe right now we only have bitbucket
	switch config.Type {
	case "bitbucket":
		bbHandler := handler.Bitbucket{}
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		err := bbHandler.SetMeUp(config, bitbucketClient)

		if err != nil {
			log.Log().Error("could not setup bitbucket client")
			return err
		}

		err = bbHandler.Walk()
		if err != nil {
			return err
		}
	}
	configPath := cred.BuildVCSCredPath(config.Type, config.AcctName)
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}
