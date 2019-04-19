package werker

import (
	"net"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/level11consulting/ocelot/build/streaminglogs"
	"github.com/level11consulting/ocelot/build/buildmonitor"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"google.golang.org/grpc"
)

//ServeMe will start HTTP Server as needed for streaming build output by hash
func ServeMe(transportChan chan *models.Transport, conf *models.WerkerFacts, store storage.OcelotStorage, buildReaper *buildmonitor.BuildReaper) {
	// todo: defer a recovery here
	werkStream := getWerkerContext(conf, store, buildReaper)
	streamPack := streaminglogs.GetStreamPack(werkStream.store, werkStream.consul)
	werkStream.streamPack = streamPack
	ocelog.Log().Debug("saving build info channels to in memory map")
	go streamPack.ListenTransport(transportChan)
	//go streamPack.ListenBuilds(buildCtxChan, sync.Mutex{})

	// do the websocket servy thing
	ocelog.Log().Info("serving websocket on port: ", conf.ServicePort)
	muxi := mux.NewRouter()
	addHandlers(muxi, werkStream)

	//start grpc server
	ocelog.Log().Info("serving grpc streams of build data on port: ", conf.GrpcPort)
	con, err := net.Listen("tcp", ":"+conf.GrpcPort)
	if err != nil {
		ocelog.Log().Fatal("womp womp")
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	werkerServer := NewWerkerServer(werkStream)
	pb.RegisterBuildServer(grpcServer, werkerServer)
	grpc_prometheus.Register(grpcServer)
	// now that grpc_prometheus is registered, can run the http1 server
	muxi.Handle("/metrics", promhttp.Handler())
	n := ocenet.InitNegroni("werker", muxi)
	go n.Run(":" + conf.ServicePort)
	// now run the grpc server
	go grpcServer.Serve(con)
}
