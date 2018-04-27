package admin

import (
	"os"
	rt "runtime"
	"path"
	"path/filepath"

	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	models "bitbucket.org/level11consulting/ocelot/models/pb"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/common/secure_grpc"

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


//Start will kick off our grpc server so it's ready to receive requests over both grpc and http
func Start(configInstance cred.CVRemoteConfig, secure secure_grpc.SecureGrpc, serverRunsAt string, port string, httpPort string) {
	//initializes our "context" - guideOcelotServer
	//store := cred.GetOcelotStorage()
	store, err := configInstance.GetOcelotStorage()
	if err != nil {
		fmt.Println("couldn't get storage instance. error: ", err.Error())
		return
	}
	defer store.Close()
	guideOcelotServer := NewGuideOcelotServer(configInstance, deserialize.New(), cred.GetValidator(), cred.GetRepoValidator(), store)
	//gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", serveSwagger)

	gw := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = models.RegisterGuideOcelotHandlerFromEndpoint(ctx, gw, serverRunsAt, opts)
	if err != nil {
		log.IncludeErrField(err).Fatal("could not register endpoints")
	}
	mux.Handle("/", gw)
	if _, ok := os.LookupEnv("SWAGGERITUP"); ok {
		go http.ListenAndServe(":" + httpPort, allowCORS(mux))
	} else {
		go http.ListenAndServe(":" + httpPort, mux)
	}

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

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	log.Log().Infof("preflight request for %s", r.URL.Path)
	return
}

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}


func serveSwagger(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
		log.Log().Errorf("Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	log.Log().Infof("Serving %s", r.URL.Path)
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	_, filename, _, _ := rt.Caller(0)
	var dir string
	dir = filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// hack, should probably take env vars
		dir = "/swagger"
	} else {
		dir = filepath.Join(dir, "models", "pb")
	}
	os.Stat(dir)
	fmt.Println(dir)
	p = path.Join(dir, p)
	http.ServeFile(w, r, p)
}


//TODO: how to propagate error codes up????
//TODO: cast this back to MY error type and set status
func CustomErrorHandler(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	// see example here: https://github.com/mycodesmells/golang-examples/blob/master/grpc/cmd/server/main.go
	ocenet.JSONApiError(w, runtime.HTTPStatusFromCode(grpc.Code(err)), "", err)
}
