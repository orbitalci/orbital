package hookhandler

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/admin/handler"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"net/http"
)

type HookHandlerContext struct {
	RemoteConfig *cred.RemoteConfig
	Producer     *nsqpb.PbProduce
	Deserializer *deserialize.Deserializer
}

//TODO: look into all the branches that's listed inside of ocelot.yml and only build if event corresonds
//tODO: branch inside of ocelot.yml

//TODO: what data do we have to store/do we need to store?
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
	//TODO: need to check and make sure that New.Type == branch
	if shouldBuild(buildConf, repopush.Push.Changes[0].New.Name) {
		tellWerker(ctx, buildConf, hash, fullName, acctName)
	}
}


// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func PullRequest(ctx *HookHandlerContext, w http.ResponseWriter, r *http.Request) {
	pr := &pb.PullRequest{}
	if err := ctx.Deserializer.JSONToProto(r.Body, pr); err != nil {
		ocelog.IncludeErrField(err).Error("could not parse request body into pb.PullRequest")
		return
	}
	ocelog.Log().Debug(r.Body)
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

	if shouldBuild(buildConf, "") {
		tellWerker(ctx, buildConf, hash, fullName, acctName)
	} else {
		//TODO: tell db that we couldn't build
	}
}

//before we build pipeline config for werker, validate and make sure this is good candidate
func shouldBuild(buildConf *pb.BuildConfig, branch string) bool {
	for _, buildBranch := range buildConf.Branches {
		if buildBranch == branch {
			return true
		}
	}
	return false
}

//TODO: this code needs to say X repo is now being tracked
//TODO: this code will also need to store status into db
func tellWerker(ctx *HookHandlerContext, buildConf *pb.BuildConfig, hash string, fullName string, acctName string) {
	// get one-time token use for access to vault
	token, err := ctx.RemoteConfig.Vault.CreateThrowawayToken()
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to create one-time vault token")
		return
	}

	werkerTask := &pb.WerkerTask{
		VaultToken:   token,
		CheckoutHash: hash,
		BuildConf: buildConf,
	}

	go ctx.Producer.WriteProto(werkerTask, "build")
}

//TODO: state = not started = to be stored inside of postgres (db interface is gonna be inside of go-til)
//this just builds the pipeline config, worker will call NewPipeline with the pipeline config and run
//func werk(ctx *HookHandlerContext, oceConfig pb.BuildConfig, gitCommit string, fullName string, acctName string) (*res.PipelineConfig, error) {
//	jobMap := make(map[string]*res.JobConfig)
//
//	var kickOffCmd []string
//	var kickOffEnvs = make(map[string]string)
//	var buildImage string
//
//	if oceConfig.Image != "" {
//		buildImage = oceConfig.Image
//	} else if len(oceConfig.Packages) > 0 {
//		buildImage = "TODO PARSE THIS AND PUSH TO ARTIFACT REPO"
//	}
//
//	//TODO: where to store failed builds
//	if oceConfig.Build == nil {
//		return nil, errors.New("Build stage cannot be empty")
//	}
//
//	//first, let's get the codebase onto the container
//	bbCreds, err := ctx.RemoteConfig.GetCredAt(cred.ConfigPath + "/bitbucket/" + acctName, false)
//	if err != nil {
//		ocelog.IncludeErrField(err)
//	}
//
//	cfg := bbCreds["bitbucket/" + acctName]
//	//TODO: clone with creds: git clone https://username:password@github.com/username/repository.git
//	kickOffCmd = append(kickOffCmd, fmt.Sprintf("wget --user=%s --password=%s https://bitbucket.org/%s/get/%s.zip", cfg.ClientId, cfg.ClientSecret, fullName, gitCommit))
//	kickOffCmd = append(kickOffCmd, fmt.Sprintf("cd $(unzip %s.zip | awk 'NR=3 {print $2}')", gitCommit))
//
//
//	if oceConfig.Before != nil {
//		if oceConfig.Before.Script != nil {
//			kickOffCmd = append(kickOffCmd, oceConfig.Before.Script...)
//		}
//
//		//combine optional before env values if passed
//		if oceConfig.Before.Env != nil {
//			for envKey, envVal := range oceConfig.Before.Env {
//				kickOffEnvs[envKey] = envVal
//			}
//		}
//	}
//
//	if oceConfig.Build != nil {
//		if oceConfig.Build.Script != nil {
//			kickOffCmd = append(kickOffCmd, oceConfig.Build.Script...)
//		}
//
//		//combine optional before env values if passed
//		if oceConfig.Build.Env != nil {
//			for envKey, envVal := range oceConfig.Build.Env {
//				kickOffEnvs[envKey] = envVal
//			}
//		}
//	}
//
//
//
//	//create a settings.xml maven file that takes in nexus and/or something else creds
//
//	//TODO: figure out what to do about the rest of the stages
//	job := &res.JobConfig{
//		Command: strings.Join(kickOffCmd, " && "),
//		Env:     kickOffEnvs,
//		Image:   buildImage,
//	}
//
//	jobMap[gitCommit] = job
//
//	pipeConfig := &res.PipelineConfig{
//		Steps:     jobMap,
//		GlobalEnv: oceConfig.Env,
//	}
//	return pipeConfig, nil
//}

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
	bbCreds, err := ctx.RemoteConfig.GetCredAt(cred.ConfigPath+"/bitbucket/"+acctName, false)
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
