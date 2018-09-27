package hookhandler

//todo: break out signaling logic and put in signaler
import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/build_signaler/webhook"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/remote"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

var (
	hookRecieves = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ocelot_recieved_hooks",
		Help: "hooks recieved and processed by hookhandler",
		// vcs_type: bitbucket | github | etc
		// event_type: pullrequest | push
	}, []string{"vcs_type", "event_type"})
)

func init() {
	prometheus.MustRegister(hookRecieves)
}

func GetContext(sig *signal.Signaler, teller *signal.PushWerkerTeller, prTeller *webhook.PullReqWerkerTeller) *HookHandlerContext {
	return &HookHandlerContext{Signaler: sig, pTeller: teller, prTeller: prTeller}
}

//context contains long lived resources. See bottom for getters/setters
type HookHandlerContext struct {
	*signal.Signaler
	pTeller  	   signal.CommitPushWerkerTeller
	prTeller 	   signal.PRWerkerTeller
	testingHandler models.VCSHandler
}

func (hhc *HookHandlerContext) getHandler(cred *pb.VCSCreds) (models.VCSHandler, string, error) {
	if hhc.testingHandler != nil {
		return hhc.testingHandler, "", nil
	}
	return remote.GetHandler(cred)
}

// On receive of repo push, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func (hhc *HookHandlerContext) RepoPush(w http.ResponseWriter, r *http.Request, vcsType pb.SubCredType) {
	hookRecieves.WithLabelValues(vcsType.String(), "push").Inc()
	translator, err := remote.GetRemoteTranslator(vcsType)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not get translator, err: ", err)
		return
	}
	push, err := translator.TranslatePush(r.Body)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not translate to proto.message, err: ", err)
		return
	}
	cred, err := credentials.GetVcsCreds(hhc.Store, push.Repo.AcctRepo, hhc.RC)
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
		ocelog.IncludeErrField(err).WithField("hash", push.HeadCommit.Hash).WithField("acctRepo", push.Repo.AcctRepo).WithField("branch", push.Branch).Error("unable to tell werker")
	}
}

//TODO: need to pass active PR branch to validator, but gonna get RepoPush handler working first
// On receive of pull request, marshal the json to an object then build the appropriate pipeline config and put on NSQ queue.
func (hhc *HookHandlerContext) PullRequest(w http.ResponseWriter, r *http.Request, vcsType pb.SubCredType) {
	hookRecieves.WithLabelValues(vcsType.String(), "pullrequest").Inc()
	translator, err := remote.GetRemoteTranslator(vcsType)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not get translator, err: ", err)
		return
	}
	pr, err := translator.TranslatePR(r.Body)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not translate to proto.message, err: ", err)
		return
	}
	cred, err := credentials.GetVcsCreds(hhc.Store, pr.Source.Repo.AcctRepo, hhc.RC)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't get creds")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not get creds, err: ", err)
		return
	}
	handler, token, err := remote.GetHandler(cred)
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
		ocelog.IncludeErrField(err).Error("couldn't get commits for PR ")
		ocenet.JSONApiError(w, http.StatusInternalServerError, "could not commits for PR, err: ", err)
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
