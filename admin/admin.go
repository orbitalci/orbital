package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/secure_grpc"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"crypto/tls"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
)

//TODO: floe integration??? just putting this note here so we remember

//Start will kick off our grpc server so it's ready to receive requests over both grpc and http
func Start(configInstance cred.CVRemoteConfig, secure secure_grpc.SecureGrpc, serverRunsAt string, port string) {
	//initializes our "context" - guideOcelotServer
	guideOcelotServer := NewGuideOcelotServer(configInstance, deserialize.New(), GetValidator(), GetRepoValidator(),
		storage.NewFileBuildStorage(""))

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

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss models.GuideOcelotServer, config *models.VCSCreds) error {
	gos := gosss.(*guideOcelotServer)

	//hehe right now we only have bitbucket
	switch config.Type {
	case "bitbucket":
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		bbHandler := handler.GetBitbucketHandler(config, bitbucketClient)
		err := bbHandler.Walk()
		if err != nil {
			return err
		}
	}
	configPath := config.BuildCredPath(config.Type, config.AcctName)
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}

func SetupRepoCredentials(gosss models.GuideOcelotServer, config *models.RepoCreds) error {
	// todo: probably should do some kind of test f they are valid or not? is there a way to test these creds
	gos := gosss.(*guideOcelotServer)
	configPath := config.BuildCredPath(config.Type, config.AcctName)
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}