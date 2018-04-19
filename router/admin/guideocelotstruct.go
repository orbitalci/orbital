package admin

import (
	"context"

	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/nsqpb"

	"bitbucket.org/level11consulting/ocelot/build"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
	"github.com/golang/protobuf/ptypes/empty"
)

//this is our grpc server, it responds to client requests
type guideOcelotServer struct {
	RemoteConfig   cred.CVRemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *cred.AdminValidator
	RepoValidator  *cred.RepoValidator
	OcyValidator   *build.OcelotValidator
	Storage        storage.OcelotStorage
	Producer       *nsqpb.PbProduce
}


// for checking if the server is reachable
func (g *guideOcelotServer) CheckConn(ctx context.Context, msg *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func NewGuideOcelotServer(config cred.CVRemoteConfig, d *deserialize.Deserializer, adminV *cred.AdminValidator, repoV *cred.RepoValidator, storage storage.OcelotStorage) pb.GuideOcelotServer {
	// changing to this style of instantiation cuz thread safe (idk read it on some best practices, it just looks
	// purdier to me anyway
	guideOcelotServer := &guideOcelotServer{
		OcyValidator:   build.GetOcelotValidator(),
		RemoteConfig:   config,
		Deserializer:   d,
		AdminValidator: adminV,
		RepoValidator:  repoV,
		Storage:        storage,
		Producer:       nsqpb.GetInitProducer(),
	}
	return guideOcelotServer
}

