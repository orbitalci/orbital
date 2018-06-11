package hookhandler

//todo: break out signaling logic and put in signaler
import (
	"net/http"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/remote"
	"github.com/shankj3/ocelot/models/pb"

)

func GetContext(sig *signal.Signaler, teller *signal.CCWerkerTeller) *HookHandlerContext {
	return &HookHandlerContext{Signaler:sig, teller:teller}
}

//context contains long lived resources. See bottom for getters/setters
type HookHandlerContext struct {
	*signal.Signaler
	// todo: CHANGE THIS
	teller   *signal.CCWerkerTeller
}


// On receive of repo push, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func (hhc *HookHandlerContext) RepoPush(w http.ResponseWriter, r *http.Request, vcsType pb.SubCredType) {
	translator, err := remote.GetRemoteTranslator(vcsType)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not get translator, err: ", err)
	}
	push, err := translator.TranslatePush(r.Body)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not translate to proto.message, err: ", err)
	}
	cred, err := credentials.GetVcsCreds(hhc.Store, push.Repo.AcctRepo, hhc.RC)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get creds")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get creds, err: ", err)
	}
	handler, token, err := remote.GetHandler(cred)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get vcs handler")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get vcs handler, err: ", err)
	}
	if err := hhc.teller.TellWerker(push.HeadCommit.Hash, hhc.Signaler, push.Branch, handler, token, push.Repo.AcctRepo, push.Commits, false, pb.SignaledBy_PUSH, nil); err != nil {
		ocelog.IncludeErrField(err).WithField("hash", push.HeadCommit.Hash).WithField("acctRepo", push.Repo.AcctRepo).WithField("branch", push.Branch).Error("unable to tell werker")
	}
}

//TODO: need to pass active PR branch to validator, but gonna get RepoPush handler working first
// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func (hhc *HookHandlerContext)  PullRequest(w http.ResponseWriter, r *http.Request, vcsType pb.SubCredType) {
	translator, err := remote.GetRemoteTranslator(vcsType)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not get translator, err: ", err)
	}
	pr, err := translator.TranslatePR(r.Body)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not translate to proto.message, err: ", err)
	}
	cred, err := credentials.GetVcsCreds(hhc.Store, pr.Source.Repo.AcctRepo, hhc.RC)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get creds")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get creds, err: ", err)
	}
	handler, token, err := remote.GetHandler(cred)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get vcs handler")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get vcs handler, err: ", err)
	}
	commits, err := handler.GetPRCommits(pr.Urls.Commits)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get commits for PR ")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not commits for PR, err: ", err)
	}

	if err := hhc.teller.TellWerker(pr.Source.Hash, hhc.Signaler, pr.Source.Branch, handler, token, pr.Source.Repo.AcctRepo, commits, false, pb.SignaledBy_PULL_REQUEST, pr); err != nil {
		ocelog.IncludeErrField(err).WithField("hash", pr.Source.Hash).WithField("acctRepo", pr.Source.Repo.AcctRepo).WithField("branch", pr.Source.Branch).Error("unable to tell werker")
	}
}

func (hhc *HookHandlerContext) HandleBBEvent(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("X-Event-Key") {
	case "repo:push":
		hhc.RepoPush(w, r, pb.SubCredType_BITBUCKET)
	case "pullrequest:created",
		"pullrequest:updated":
		hhc.PullRequest(w, r, pb.SubCredType_BITBUCKET)
	default:
		ocelog.Log().Errorf("No support for Bitbucket event %s", r.Header.Get("X-Event-Key"))
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}
