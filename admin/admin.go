package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/secure_grpc"

	//"bitbucket.org/level11consulting/ocelot/util/handler"
	//"bitbucket.org/level11consulting/ocelot/util/secure_grpc"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
)

//TODO: floe integration? putting this note here so we remember

//con, err := net.Listen("tcp", ":"+conf.GrpcPort)
//if err != nil {
//ocelog.Log().Fatal("womp womp")
//}
//grpcServer := grpc.NewServer()
//protobuf.RegisterBuildServer(grpcServer, werkStream)
//go grpcServer.Serve(con)

//Start will kick off our grpc server so it's ready to receive requests over both grpc and http
func Start(configInstance cred.CVRemoteConfig, secure secure_grpc.SecureGrpc, serverRunsAt string, port string) {
	//initializes our "context" - guideOcelotServer
	//store := cred.GetOcelotStorage()
	store, err := configInstance.GetOcelotStorage()
	if err != nil {
		fmt.Println("couldn't get storage instance. error: ", err.Error())
		return
	}
	guideOcelotServer := NewGuideOcelotServer(configInstance, deserialize.New(), GetValidator(), GetRepoValidator(), store)
	//grpc server
	con, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Log().Fatal("listen: ", err)
	}
	grpcServer := grpc.NewServer()
	models.RegisterGuideOcelotServer(grpcServer, guideOcelotServer)
	err = grpcServer.Serve(con)
	if err != nil {
		log.Log().Fatal("serve: ", err)
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
