package main

import (
	"github.com/gorilla/mux"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	pb "github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util"
	"github.com/shankj3/ocelot/util/deserialize"
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	"leveler/resources"
	"net/http"
	"os"
	"strings"
)

type HookHandlerContext struct {
	RemoteConfig *util.RemoteConfig
	Producer     *nsqpb.PbProduce
	Deserializer *deserialize.Deserializer
}

// On receive of repo push, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func RepoPush(ctx *HookHandlerContext, w http.ResponseWriter, r *http.Request) {
	repopush := &pb.RepoPush{}
	if err := ctx.Deserializer.JSONToProto(r.Body, repopush); err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not parse request body into proto.Message", err)
	}
	fullName := repopush.Repository.FullName
	hash := repopush.Push.Changes[0].New.Target.Hash
	acctName := repopush.Repository.Owner.Username
	buildConf, err := GetBBBuildConfig(ctx, acctName, fullName, hash)
	if err != nil {
		// if the build file just isn't there don't worry about it.
		if err != ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("unable to get build conf")
			return
		}
		ocelog.Log().Debugf("no ocelot yml found for repo %s", repopush.Repository.FullName)
		return
	}
	tellWerker(ctx, buildConf, hash)
}

// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func PullRequest(ctx *HookHandlerContext, w http.ResponseWriter, r *http.Request) {
	pr := &pb.PullRequest{}
	if err := ctx.Deserializer.JSONToProto(r.Body, pr); err != nil {
		ocelog.IncludeErrField(err).Error("could not parse request body into pb.PullRequest")
		return
	}
	fullName := pr.Pullrequest.Source.Repository.FullName
	hash := pr.Pullrequest.Source.Commit.Hash
	acctName := pr.Pullrequest.Source.Repository.Owner.Username

	buildConf, err := GetBBBuildConfig(ctx, acctName, fullName, hash)
	if err != nil {
		// if the build file just isn't there don't worry about it.
		if err != ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("unable to get build conf")
			return
		}
		ocelog.Log().Debugf("no ocelot yml found for repo %s", pr.Pullrequest.Source.Repository.FullName)
		return
	}
	tellWerker(ctx, buildConf, hash)
}

func tellWerker(ctx *HookHandlerContext, buildConf *pb.BuildConfig, hash string) {
	// get one-time token use for access to vault
	token, err := ctx.RemoteConfig.Vault.CreateThrowawayToken()
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to create one-time vault token")
		return
	}

	pipeConfig, err := werk(*buildConf, hash)

	werkerTask := WerkerTask{
		VaultToken:   token,
		CheckoutHash: hash,
		Pipe:         pipeConfig,
	}

	go ctx.Producer.WriteAny(werkerTask, nsqpb.Build)
}

//TODO: this needs to live not in this file
type WerkerTask struct {
	VaultToken   string
	CheckoutHash string
	Pipe         *resources.Pipeline
}

//this just builds the pipeline config, worker will call NewPipeline with the protobuf pipeline config and run
func werk(oceConfig pb.BuildConfig, gitCommit string) (*resources.Pipeline, error) {
	//TODO: avoid passing integration (this is sensitive data) around by not setting it, maybe mistake?
	// integration will be list of strings we can parse on werker side and it will do things with it there
	//TODO: what should be the key in this job map?
	// key = job name, and needs to be unique within the pipeline
	//TODO: example input for job? What should be passed ot list of strings?
	// inputs/outputs in a JOB are the keys to pipeline input/outputs in PipelineConfig
	//TODO: potentially think about how to use integration?
	//TODO: how/when do we push artifacts to nexus? (think about this while I'm writing other code)
		// TODO: potentially watch for changes in .m2/PKG_NAME with fsnotify?
	//TODO: we might be able to actually create an image and use input/outputs for the dockerPackages part
	//TODO: seems like image and dockerPackages should be mutually exclusive?
		// reasoning: can handle weird specific docker package needs (like downloading maven version CANNOT just be apt-get)
		// would make my life easier

	jobMap := make(map[string]*resources.Job)

	var kickOffCmd []string
	var kickOffEnvs = make(map[string]string)
	var buildImage string

	if oceConfig.Image != "" {
		buildImage = oceConfig.Image
	} else if len(oceConfig.DockerPackages) > 0 {
		buildImage = "TODO PARSE THIS AND PUSH TO NEXUS"
		//TODO: build image and store it somewhere. OH! NEXUS! oh shit we need nexus int. now
	}

	if oceConfig.Before != nil {
		if oceConfig.Before.Script != nil {
			kickOffCmd = append(kickOffCmd, oceConfig.Before.Script...)
		}

		//combine optional before env values if passed
		if oceConfig.Before.Env != nil {
			for envKey, envVal := range oceConfig.Before.Env {
				kickOffEnvs[envKey] = envVal
			}
		}
	}

	if oceConfig.Build != nil {
		if oceConfig.Build.Script != nil {
			kickOffCmd = append(kickOffCmd, oceConfig.Build.Script...)
		}

		//combine optional before env values if passed
		if oceConfig.Build.Env != nil {
			for envKey, envVal := range oceConfig.Build.Env {
				kickOffEnvs[envKey] = envVal
			}
		}
	}

	//TODO: figure out what to do about the rest of the stages

	job := &resources.Job{
		Command: strings.Join(kickOffCmd, " && "),
		Env:     kickOffEnvs,
		Image:   buildImage,
	}

	jobMap[gitCommit] = job

	pipeConfig := &resources.Pipeline{
		Steps: jobMap,
		GlobalEnv: oceConfig.Env,
	}
	return pipeConfig, nil
}

func HandleBBEvent(ctx interface{}, w http.ResponseWriter, r *http.Request) {
	handlerCtx := ctx.(*HookHandlerContext)

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
func getCredConfig() *models.Credentials {
	return &models.Credentials{
		ClientId:     "QEBYwP5cKAC3ykhau4",
		ClientSecret: "gKY2S3NGnFzJKBtUTGjQKc4UNvQqa2Vb",
		TokenURL:     "https://bitbucket.org/site/oauth2/access_token",
		AcctName:     "jessishank",
	}
}

func GetBBBuildConfig(ctx *HookHandlerContext, acctName string, repoFullName string, checkoutCommit string) (conf *pb.BuildConfig, err error) {
	//cfg := getCredConfig()
	bbCreds, err := ctx.RemoteConfig.GetCredAt(util.ConfigPath+"/bitbucket/"+acctName, false)
	cfg := bbCreds["bitbucket/"+acctName]
	bb := handler.Bitbucket{}
	bbClient := &ocenet.OAuthClient{}
	bbClient.Setup(cfg)

	bb.SetMeUp(cfg, bbClient)
	fileBitz, err := bb.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		return
	}
	conf = &pb.BuildConfig{}
	if err != nil {
		return
	}
	if err = ctx.Deserializer.YAMLToStruct(fileBitz, conf); err != nil {
		return
	}
	return
}

//TODO: move this so that the kickoff starts in cmd, and edit the dockerfiles accordingly
func main() {
	ocelog.InitializeOcelog(ocelog.GetFlags())
	ocelog.Log().Debug()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
		ocelog.Log().Warning("running on default port :8088")
	}

	remoteConfig, err := util.GetInstance("", 0, "")
	if err != nil {
		ocelog.Log().Fatal(err)
	}

	hookHandlerContext := &HookHandlerContext{
		RemoteConfig: remoteConfig,
		Deserializer: deserialize.New(),
		Producer:     nsqpb.GetInitProducer(),
	}

	muxi := mux.NewRouter()

	// handleBBevent can take push/pull/ w/e
	muxi.Handle("/bitbucket", &ocenet.AppContextHandler{hookHandlerContext, HandleBBEvent}).Methods("POST")

	// mux.HandleFunc("/", ViewWebhooks).Methods("GET")
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}
