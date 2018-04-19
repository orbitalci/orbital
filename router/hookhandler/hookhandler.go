package hookhandler
//todo: break out signaling logic and put in signaler
import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/build"
	signal "bitbucket.org/level11consulting/ocelot/build_signaler"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	pbb "bitbucket.org/level11consulting/ocelot/models/bitbucket/pb"
	"bitbucket.org/level11consulting/ocelot/storage"

	"net/http"
)


type HookHandler interface {
	GetRemoteConfig() cred.CVRemoteConfig
	SetRemoteConfig(remoteConfig cred.CVRemoteConfig)
	GetProducer() *nsqpb.PbProduce
	SetProducer(producer *nsqpb.PbProduce)
	GetDeserializer() *deserialize.Deserializer
	SetDeserializer(deserializer *deserialize.Deserializer)
	GetValidator() *build.OcelotValidator
	SetValidator(validator *build.OcelotValidator)
	GetStorage() storage.OcelotStorage
	SetStorage(storage.OcelotStorage)
	GetTeller() *signal.BBWerkerTeller
	GetSignaler() *signal.Signaler
}

//context contains long lived resources. See bottom for getters/setters
type HookHandlerContext struct {
 	*signal.Signaler
 	teller *signal.BBWerkerTeller
}

// On receive of repo push, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func RepoPush(ctx HookHandler, w http.ResponseWriter, r *http.Request) {
	repopush := &pbb.RepoPush{}

	if err := ctx.GetDeserializer().JSONToProto(r.Body, repopush); err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not parse request body into proto.Message", err)
	}

	fullName := repopush.Repository.FullName
	hash := repopush.Push.Changes[0].New.Target.Hash
	branch := repopush.Push.Changes[0].New.Name
	//acctName := repopush.Repository.Owner.Username

	if err := ctx.GetTeller().TellWerker(hash, ctx.GetSignaler(), branch, nil, "" ); err != nil {
		ocelog.IncludeErrField(err).WithField("hash", hash).WithField("acctRepo", fullName).WithField("branch", branch).Error("unable to tell werker")
	}
}


//TODO: need to pass active PR branch to validator, but gonna get RepoPush handler working first
// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func PullRequest(ctx HookHandler, w http.ResponseWriter, r *http.Request) {
	pr := &pbb.PullRequest{}
	if err := ctx.GetDeserializer().JSONToProto(r.Body, pr); err != nil {
		ocelog.IncludeErrField(err).Error("could not parse request body into pb.PullRequest")
		return
	}
	ocelog.Log().Debug(r.Body)
	fullName := pr.Pullrequest.Source.Repository.FullName
	hash := pr.Pullrequest.Source.Commit.Hash
	//acctName := pr.Pullrequest.Source.Repository.Owner.Username
	branch := pr.Pullrequest.Source.Branch.Name

	if err := ctx.GetTeller().TellWerker(hash, ctx.GetSignaler(), branch, nil, "" ); err != nil {
		ocelog.IncludeErrField(err).WithField("hash", hash).WithField("acctRepo", fullName).WithField("branch", branch).Error("unable to tell werker")
	}
}

func HandleBBEvent(ctx interface{}, w http.ResponseWriter, r *http.Request) {
	handlerCtx := ctx.(HookHandler)

	switch r.Header.Get("X-Event-Key") {
	case "repo:push":
		RepoPush(handlerCtx, w, r)
	case "pullrequest:created",
		"pullrequest:updated":
		PullRequest(handlerCtx, w, r)
	default:
		ocelog.Log().Errorf("No support for Bitbucket event %s", r.Header.Get("X-Event-Key"))
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (hhc *HookHandlerContext) GetRemoteConfig() cred.CVRemoteConfig {
	return hhc.RC
}
func (hhc *HookHandlerContext) SetRemoteConfig(remoteConfig cred.CVRemoteConfig) {
	hhc.RC = remoteConfig
}
func (hhc *HookHandlerContext) GetProducer() *nsqpb.PbProduce {
	return hhc.Producer
}
func (hhc *HookHandlerContext) SetProducer(producer *nsqpb.PbProduce) {
	hhc.Producer = producer
}
func (hhc *HookHandlerContext) GetDeserializer() *deserialize.Deserializer {
	return hhc.Deserializer
}
func (hhc *HookHandlerContext) SetDeserializer(deserializer *deserialize.Deserializer) {
	hhc.Deserializer = deserializer
}
func (hhc *HookHandlerContext) SetValidator(validator *build.OcelotValidator) {
	hhc.OcyValidator = validator
}

func (hhc *HookHandlerContext) GetValidator() *build.OcelotValidator {
	return hhc.OcyValidator
}

func (hhc *HookHandlerContext) SetStorage(ocelotStorage storage.OcelotStorage) {
	hhc.Store = ocelotStorage
}

func (hhc *HookHandlerContext) GetStorage() storage.OcelotStorage {
	return hhc.Store
}

func (hhc *HookHandlerContext) GetTeller() *signal.BBWerkerTeller {
	return hhc.teller
}

func (hhc *HookHandlerContext) SetTeller(tell *signal.BBWerkerTeller) {
	hhc.teller = tell
}

func (hhc *HookHandlerContext) GetSignaler() *signal.Signaler {
	return hhc.Signaler
}
