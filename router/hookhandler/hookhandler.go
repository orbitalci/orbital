package hookhandler

//todo: break out signaling logic and put in signaler
import (
	"fmt"
	"net/http"
	"strings"

	"github.com/level11consulting/ocelot/build/vcshandler"
	signal "github.com/level11consulting/ocelot/build_signaler"
	"github.com/level11consulting/ocelot/build_signaler/webhook"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
)

func GetContext(sig *signal.Signaler, teller *signal.PushWerkerTeller, prTeller *webhook.PullReqWerkerTeller) *HookHandlerContext {
	return &HookHandlerContext{Signaler: sig, pTeller: teller, prTeller: prTeller}
}

//HookHandlerContext contains long lived resources. See bottom for getters/setters
type HookHandlerContext struct {
	*signal.Signaler
	pTeller        signal.CommitPushWerkerTeller
	prTeller       signal.PRWerkerTeller
	testingHandler models.VCSHandler
}

func (hhc *HookHandlerContext) getHandler(cred *pb.VCSCreds) (models.VCSHandler, string, error) {
	if hhc.testingHandler != nil {
		return hhc.testingHandler, "token", nil
	}
	return vcshandler.GetHandler(cred)
}

// On receive of repo push, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func (hhc *HookHandlerContext) RepoPush(w http.ResponseWriter, r *http.Request, vcsType pb.SubCredType) {
	hookRecieves.WithLabelValues(vcsType.String(), "push").Inc()
	translator, err := vcshandler.GetRemoteTranslator(vcsType)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not get translator, err: ", err)
		return
	}
	push, err := translator.TranslatePush(r.Body)
	if err != nil {
		if _, ok := err.(*models.DontBuildThisEvent); ok {
			ocelog.Log().Infof("not building event because the translator noticed it is not viable: %s", err.Error())
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		failedTranslation.WithLabelValues("push").Inc()
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not translate to proto.message, err: ", err)
		return
	}
	if strings.ToLower(ocelog.GetLogLevel().String()) == "debug" {
		var commits string
		for _, commit := range push.Commits {
			commits += fmt.Sprintf("%s:%s\n", commit.GetHash(), commit.GetMessage())
		}
		ocelog.Log().Infof("NEW COMMITS ARE IN! %s", commits)
	}
	cred, err := config.GetVcsCreds(hhc.Store, push.Repo.AcctRepo, hhc.RC, vcsType)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get creds")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get creds, err: ", err)
		return
	}
	handler, token, err := hhc.getHandler(cred)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get vcs handler")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get vcs handler, err: ", err)
		return
	}
	if err := hhc.pTeller.TellWerker(push, hhc.Signaler, handler, token, false, pb.SignaledBy_PUSH); err != nil {
		failedToTellWerker.Inc()
		ocelog.IncludeErrField(err).WithField("hash", push.HeadCommit.Hash).WithField("acctRepo", push.Repo.AcctRepo).WithField("branch", push.Branch).Error("unable to tell werker")
		return
	}
	ocelog.Log().WithField("hash", push.GetHeadCommit().GetHash()).WithField("acctRepo", push.GetRepo().GetAcctRepo()).Info("succesfully queued build from push event")
}

// TODO: need to pass active PR branch to validator, but gonna get RepoPush handler working first
// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func (hhc *HookHandlerContext) PullRequest(w http.ResponseWriter, r *http.Request, vcsType pb.SubCredType) {
	hookRecieves.WithLabelValues(vcsType.String(), "pullrequest").Inc()
	translator, err := vcshandler.GetRemoteTranslator(vcsType)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not get translator, err: ", err)
		return
	}
	pr, err := translator.TranslatePR(r.Body)
	if err != nil {
		if _, ok := err.(*models.DontBuildThisEvent); ok {
			ocelog.Log().Infof("not building event because the translator noticed it is not viable: %s", err.Error())
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		failedTranslation.WithLabelValues("pullrequest").Inc()
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not translate to proto.message, err: ", err)
		return
	}
	cred, err := config.GetVcsCreds(hhc.Store, pr.Source.Repo.AcctRepo, hhc.RC, vcsType)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get creds")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get creds, err: ", err)
		return
	}
	handler, token, err := hhc.getHandler(cred)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get vcs handler")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get vcs handler, err: ", err)
		return
	}
	prData := &pb.PrWerkerData{
		Urls: &pb.PrUrls{
			Approve:  pr.Urls.Approve,
			Decline:  pr.Urls.Decline,
			Comments: pr.Urls.Comments,
			Commits:  pr.Urls.Commits,
			Statuses: pr.Urls.Statuses,
			Merge:    pr.Urls.Merge,
		},
		PrId: fmt.Sprintf("%d", pr.Id),
	}
	if err = hhc.prTeller.TellWerker(pr, prData, hhc.Signaler, handler, token, false, pb.SignaledBy_PULL_REQUEST); err != nil {
		failedToTellWerker.Inc()
		ocelog.IncludeErrField(err).Error("couldn't get commits for PR ")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not commits for PR, err: ", err)
	}
	ocelog.Log().WithField("hash", pr.GetSource().GetHash()).WithField("acctRepo", pr.GetSource().GetRepo().GetAcctRepo()).Info("succesfully queued build from PR event")
}

func (hhc *HookHandlerContext) HandleBBEvent(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("X-Event-Key") {
	case "repo:push":
		hhc.RepoPush(w, r, pb.SubCredType_BITBUCKET)
	case "pullrequest:created",
		"pullrequest:updated":
		hhc.PullRequest(w, r, pb.SubCredType_BITBUCKET)
	default:
		unprocessibleEvent.WithLabelValues(r.Header.Get("X-Event-Key"), pb.SubCredType_BITBUCKET.String())
		ocelog.Log().Errorf("No support for Bitbucket event %s", r.Header.Get("X-Event-Key"))
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (hhc *HookHandlerContext) HandleGHEvent(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("X-GitHub-Event") {
	case "push":
		hhc.RepoPush(w, r, pb.SubCredType_GITHUB)
	case "pull_request":
		hhc.PullRequest(w, r, pb.SubCredType_GITHUB)
	default:
		unprocessibleEvent.WithLabelValues(r.Header.Get("X-GitHub-Event"), pb.SubCredType_GITHUB.String())
		ocelog.Log().Errorf("No support for Github event %s", r.Header.Get("X-GitHub-Event"))
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}
