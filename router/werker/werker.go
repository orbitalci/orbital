package werker

import (
	"net"
	"net/http"

	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/build/streamer"
	"bitbucket.org/level11consulting/ocelot/build/valet"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

//ServeMe will start HTTP Server as needed for streaming build output by hash
func ServeMe(transportChan chan *models.Transport, conf *models.WerkerFacts, store storage.OcelotStorage, killValet *valet.ContextValet) {
	// todo: defer a recovery here
	werkStream := getWerkerContext(conf, store, killValet)
	streamPack := streamer.GetStreamPack(werkStream.store, werkStream.consul)
	werkStream.streamPack = streamPack
	ocelog.Log().Debug("saving build info channels to in memory map")
	go streamPack.ListenTransport(transportChan)
	//go streamPack.ListenBuilds(buildCtxChan, sync.Mutex{})

	// do the websocket servy thing
	ocelog.Log().Info("serving websocket on port: ", conf.ServicePort)
	muxi := mux.NewRouter()
	addHandlers(muxi, werkStream)
	n := ocenet.InitNegroni("werker", muxi)
	go n.Run(":" + conf.ServicePort)

	//start grpc server
	ocelog.Log().Info("serving grpc streams of build data on port: ", conf.GrpcPort)
	con, err := net.Listen("tcp", ":"+conf.GrpcPort)
	if err != nil {
		ocelog.Log().Fatal("womp womp")
	}

	grpcServer := grpc.NewServer()
	werkerServer := NewWerkerServer(werkStream)
	pb.RegisterBuildServer(grpcServer, werkerServer)
	go grpcServer.Serve(con)
	go func() {
		ocelog.Log().Info(http.ListenAndServe(":6060", nil))
	}()
}
