package storage

import (
	"fmt"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"time"
)

type Dest int

const (
	FileSystem Dest = iota
	Postgres
)

//go:generate mockgen -source storage.go -destination storage.mock.go -package storage

type Stringy interface {
	StorageType() string
}

type BuildOut interface {
	AddOut(output *models.BuildOutput) error
	RetrieveOut(buildId int64) (models.BuildOutput, error)
	RetrieveLastOutByHash(gitHash string) (models.BuildOutput, error)
}

type BuildSum interface {
	// AddSumStart will
	AddSumStart(hash string, account string, repo string, branch string, by pb.SignaledBy, credId int64) (int64, error)
	UpdateSum(failed bool, duration float64, id int64) error
	RetrieveSumByBuildId(buildId int64) (*pb.BuildSummary, error)
	RetrieveSum(gitHash string) ([]*pb.BuildSummary, error)
	RetrieveLatestSum(gitHash string) (*pb.BuildSummary, error)
	RetrieveHashStartsWith(partialGitHash string) ([]*pb.BuildSummary, error)
	RetrieveLastFewSums(repo string, account string, limit int32) ([]*pb.BuildSummary, error)
	RetrieveAcctRepo(partialRepo string) ([]*pb.BuildSummary, error)
	StartBuild(id int64) error
	StoreFailedValidation(id int64) error
	SetQueueTime(id int64) error
	GetTrackedRepos() (*pb.AcctRepos, error)
	GetLastSuccessfulBuildHash(account, repo, branch string) (string, error)
}

type BuildStage interface {
	AddStageDetail(stageResult *models.StageResult) error
	RetrieveStageDetail(buildId int64) ([]models.StageResult, error)
}

type PollTable interface {
	InsertPoll(account string, repo string, cronString string, branches string, credsId int64) error
	UpdatePoll(account string, repo string, cronString string, branches string) error
	SetLastData(account string, repo string, lasthashes map[string]string) error
	GetLastData(accountRepo string) (timestamp time.Time, hashes map[string]string, err error)
	PollExists(account string, repo string) (bool, error)
	GetAllPolls() ([]*pb.PollRequest, error)
	DeletePoll(account string, repo string) error
}

type CredTable interface {
	InsertCred(credder pb.OcyCredder, overWriteOk bool) error
	// retrieve ordered by cred type
	RetrieveAllCreds() ([]pb.OcyCredder, error)
	// todo: take out all this pb.CredType stuff, can just call Parent() on subcredType. realized too late :"(
	RetrieveCreds(credType pb.CredType) ([]pb.OcyCredder, error)
	RetrieveCred(subCredType pb.SubCredType, identifier, accountName string) (pb.OcyCredder, error)
	RetrieveCredBySubTypeAndAcct(scredType pb.SubCredType, acctName string) ([]pb.OcyCredder, error)
	CredExists(credder pb.OcyCredder) (bool, error)
	UpdateCred(credder pb.OcyCredder) error
	DeleteCred(credder pb.OcyCredder) error
	GetVCSTypeFromAccount(account string) (pb.SubCredType, error)
}

type SubscriptionsTable interface {
	InsertOrUpdateActiveSubscription(subscriptions *pb.ActiveSubscription) (int64, error)
	FindSubscribeesForRepo(acctRepo string, credType pb.SubCredType) ([]*pb.ActiveSubscription, error)
	GetSubscriptionData(subscribingAcctRepo string, subscribingBuildId int64, subscribingVcsType pb.SubCredType) (data *pb.SubscriptionUpstreamData, err error)
}

//GetCredAt(path string, hideSecret bool, rcc RemoteConfigCred) (map[string]RemoteConfigCred, error)
type HealthyChkr interface {
	Healthy() bool
}

type OcelotStorage interface {
	BuildOut
	BuildSum
	BuildStage
	Stringy
	PollTable
	CredTable
	SubscriptionsTable
	HealthyChkr
	Close()
}

var (
	BUILD_SUM_404    = "no build summary found for %s"
	STAGE_REASON_404 = "no stages found for %s"
	BUILD_OUT_404    = "no build output found for %s"
	CRED_404         = "no credential found for %s %s"
)

func BuildSumNotFound(id string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(BUILD_SUM_404, id)}
}

func BuildOutNotFound(id string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(BUILD_OUT_404, id)}
}

func StagesNotFound(id string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(STAGE_REASON_404, id)}
}

func CredNotFound(account string, repoType string) *ErrNotFound {
	return &ErrNotFound{fmt.Sprintf(CRED_404, account, repoType)}
}

type ErrNotFound struct {
	msg string
}

func (e *ErrNotFound) Error() string {
	return e.msg
}

func MultipleVCSTypes(account string, types []pb.SubCredType) *ErrMultipleVCSTypes {
	return &ErrMultipleVCSTypes{account: account, types: types}
}

type ErrMultipleVCSTypes struct {
	account string
	types []pb.SubCredType
}

func (e *ErrMultipleVCSTypes) Error() string {
	var stringTypes []string
	for _, sct := range e.types {
		stringTypes = append(stringTypes, sct.String())
	}
	return fmt.Sprintf("there are multiple vcs types to the account %s: %s", e.account, stringTypes)
}