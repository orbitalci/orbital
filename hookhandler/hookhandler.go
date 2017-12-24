package hookhandler

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"fmt"
	"github.com/go-errors/errors"
	"net/http"
)


type HookHandler interface {
	GetBitbucketClient(cfg *models.VCSCreds) (handler.VCSHandler, string, error)
	GetRemoteConfig() *cred.RemoteConfig
	SetRemoteConfig(remoteConfig *cred.RemoteConfig)
	GetProducer() *nsqpb.PbProduce
	SetProducer(producer *nsqpb.PbProduce)
	GetDeserializer() *deserialize.Deserializer
	SetDeserializer(deserializer *deserialize.Deserializer)
}


type HookHandlerContext struct {
	RemoteConfig *cred.RemoteConfig
	Producer     *nsqpb.PbProduce
	Deserializer *deserialize.Deserializer
}

//Returns VCS handler for pulling source code and auth token if exists (auth token is needed for code download)
func (hhc *HookHandlerContext) GetBitbucketClient(cfg *models.VCSCreds) (handler.VCSHandler, string, error) {
	bbClient := &ocenet.OAuthClient{}
	token, err := bbClient.Setup(cfg)
	if err != nil {
		return nil, "", err
	}
	bb := handler.GetBitbucketHandler(cfg, bbClient)
	return bb, token, nil
}

func (hhc *HookHandlerContext) GetRemoteConfig() *cred.RemoteConfig {
	return hhc.RemoteConfig
}
func (hhc *HookHandlerContext) SetRemoteConfig(remoteConfig *cred.RemoteConfig) {
	hhc.RemoteConfig = remoteConfig
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


// On receive of repo push, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func RepoPush(ctx HookHandler, w http.ResponseWriter, r *http.Request) {
	repopush := &pb.RepoPush{}

	if err := ctx.GetDeserializer().JSONToProto(r.Body, repopush); err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not parse request body into proto.Message", err)
	}

	fullName := repopush.Repository.FullName
	hash := repopush.Push.Changes[0].New.Target.Hash
	acctName := repopush.Repository.Owner.Username
	buildConf, bbToken, err := GetBBConfig(ctx, acctName, fullName, hash)
	if err != nil {
		// if the build file just isn't there don't worry about it.
		if err != ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("unable to get build conf")
			return
		}
		ocelog.Log().Debugf("no ocelot yml found for repo %s", repopush.Repository.FullName)
		return
	}
	fmt.Println(buildConf)
	//TODO: need to check and make sure that New.Type == branch
	if validateBuild(buildConf, repopush.Push.Changes[0].New.Name) {
		tellWerker(ctx, buildConf, hash, fullName, bbToken)
	} else {
		//TODO: tell db we couldn't build
	}
}


//TODO: need to pass active PR branch to validator, but gonna get RepoPush handler working first
// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func PullRequest(ctx HookHandler, w http.ResponseWriter, r *http.Request) {
	pr := &pb.PullRequest{}
	if err := ctx.GetDeserializer().JSONToProto(r.Body, pr); err != nil {
		ocelog.IncludeErrField(err).Error("could not parse request body into pb.PullRequest")
		return
	}
	ocelog.Log().Debug(r.Body)
	fullName := pr.Pullrequest.Source.Repository.FullName
	hash := pr.Pullrequest.Source.Commit.Hash
	acctName := pr.Pullrequest.Source.Repository.Owner.Username

	buildConf, bbToken, err := GetBBConfig(ctx, acctName, fullName, hash)
	if err != nil {
		// if the build file just isn't there don't worry about it.
		if err != ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("unable to get build conf")
			return
		}
		ocelog.Log().Debugf("no ocelot yml found for repo %s", pr.Pullrequest.Source.Repository.FullName)
		return
	}

	if validateBuild(buildConf, "") {
		tellWerker(ctx, buildConf, hash, fullName, bbToken)
	} else {
		//TODO: tell db we couldn't build
	}
}

//before we build pipeline config for werker, validate and make sure this is good candidate
	// - check if commit branch matches with ocelot.yaml branch
	// - check if ocelot.yaml has at least one step called build
//TODO: move validator out to its own class and whatnot, that way admin or command line client can use to validate
func validateBuild(buildConf *pb.BuildConfig, branch string) bool {
	_, ok := buildConf.Stages["build"]
	if !ok {
		ocelog.Log().Error("your ocelot.yml does not have the required `build` stage")
		return false
	}

	for _, buildBranch := range buildConf.Branches {
		if buildBranch == branch {
			return true
		}
	}
	ocelog.Log().Errorf("build does not match any branches listed: %v", buildConf.Branches)
	return false
}


//TODO: this code needs to store status into db
func tellWerker(ctx HookHandler, buildConf *pb.BuildConfig, hash string, fullName string, bbToken string) {
	// get one-time token use for access to vault
	token, err := ctx.GetRemoteConfig().Vault.CreateThrowawayToken()
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to create one-time vault token")
		return
	}

	werkerTask := &pb.WerkerTask{
		VaultToken:   token,
		CheckoutHash: hash,
		BuildConf: buildConf,
		VcsToken: bbToken,
		VcsType: "bitbucket",
		FullName: fullName,
	}

	go ctx.GetProducer().WriteProto(werkerTask, "build")
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

// for testing
func getCredConfig() *models.VCSCreds {
	return &models.VCSCreds{
		ClientId:     "QEBYwP5cKAC3ykhau4",
		ClientSecret: "gKY2S3NGnFzJKBtUTGjQKc4UNvQqa2Vb",
		TokenURL:     "https://bitbucket.org/site/oauth2/access_token",
		AcctName:     "jessishank",
	}
}

//returns config if it exists, bitbucket token, and err
func GetBBConfig(ctx HookHandler, acctName string, repoFullName string, checkoutCommit string) (conf *pb.BuildConfig, token string, err error) {
	bbCreds, err := ctx.GetRemoteConfig().GetCredAt(cred.BuildCredPath("bitbucket", acctName, cred.Vcs), false, cred.Vcs)
	cf := bbCreds["bitbucket/"+acctName]
	cfg, ok := cf.(*models.VCSCreds)

	if !ok {
		err = errors.New(fmt.Sprintf("could not cast config as models.VCSCreds, config: %v", cf))
		return
	}

	bbClient := &ocenet.OAuthClient{}
	bbClient.Setup(cfg)

	bb, token, err := ctx.GetBitbucketClient(cfg)
	if err != nil {
		return
	}

	fileBitz, err := bb.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		return
	}
	conf = &pb.BuildConfig{}
	if err != nil {
		return
	}
	if err = ctx.GetDeserializer().YAMLToStruct(fileBitz, conf); err != nil {
		return
	}
	return
}
