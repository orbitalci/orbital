package hookhandler

//todo: break out signaling logic and put in signaler
import (
	"github.com/shankj3/go-til/deserialize"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/build"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/build_signaler/poll"
	cred "github.com/shankj3/ocelot/common/credentials"
	pbb "github.com/shankj3/ocelot/models/bitbucket/pb"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"

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
	// todo: CHANGE THIS
	GetTeller() *poll.CCWerkerTeller
	GetSignaler() *signal.Signaler
}

//context contains long lived resources. See bottom for getters/setters
type HookHandlerContext struct {
	*signal.Signaler
	// todo: CHANGE THIS
	teller *poll.CCWerkerTeller
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
	var commits []*pb.Commit
	for _, commit := range repopush.Push.Changes[0].Commits {
		commits = append(commits, &pb.Commit{Hash:commit.Hash, Date: commit.Date, Message:commit.Message})
	}
	checkData := build.NewViable(branch, nil, commits, false)
	if err := ctx.GetTeller().TellWerker(hash, ctx.GetSignaler(), branch, nil, "", fullName, checkData); err != nil {
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
	var commits []*pb.Commit
	//for _, commit := range pr.Pullrequest.
	// todo: hit an endpoint w/ vcs to get all commits of a pull request
	// TODO-continued: bitbucket example: https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Busername%7D/%7Brepo_slug%7D/pullrequests/%7Bpull_request_id%7D/commits
	checkData := build.NewViable(branch, nil, commits, false)
	if err := ctx.GetTeller().TellWerker(hash, ctx.GetSignaler(), branch, nil, "", fullName, checkData); err != nil {
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

func (hhc *HookHandlerContext) GetTeller() *poll.CCWerkerTeller {
	return hhc.teller
}

func (hhc *HookHandlerContext) SetTeller(tell *poll.CCWerkerTeller) {
	hhc.teller = tell
}

func (hhc *HookHandlerContext) GetSignaler() *signal.Signaler {
	return hhc.Signaler
}
