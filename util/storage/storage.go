package storage

import (
	pb "bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"fmt"
	"time"
)

type Dest int

const (
	FileSystem Dest = iota
	Postgres
)

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
	AddSumStart(hash string, account string, repo string, branch string) (int64, error)
	UpdateSum(failed bool, duration float64, id int64) error
	RetrieveSumByBuildId(buildId int64) (models.BuildSummary, error)
	RetrieveSum(gitHash string) ([]models.BuildSummary, error)
	RetrieveLatestSum(gitHash string) (models.BuildSummary, error)
	RetrieveHashStartsWith(partialGitHash string) ([]models.BuildSummary, error)
	RetrieveLastFewSums(repo string, account string, limit int32) ([]models.BuildSummary, error)
	RetrieveAcctRepo(partialRepo string) ([]models.BuildSummary, error)
	StartBuild(id int64) error
	StoreFailedValidation(id int64) error
	SetQueueTime(id int64) error

}

type BuildStage interface {
	AddStageDetail(stageResult *models.StageResult) error
	RetrieveStageDetail(buildId int64) ([]models.StageResult, error)
}

type PollTable interface {
	InsertPoll(account string, repo string, cronString string, branches string) error
	UpdatePoll(account string, repo string, cronString string, branches string) error
	SetLastData(account string, repo string, lasthashes map[string]string) error
	GetLastData(accountRepo string) (timestamp time.Time, hashes map[string]string, err error)
	PollExists(account string, repo string) (bool, error)
	GetAllPolls() ([]*models.PollRequest, error)
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
	HealthyChkr
	Close()
}

var (
	BUILD_SUM_404 = "no build summary found for %s"
	STAGE_REASON_404 = "no stages found for %s"
	BUILD_OUT_404 = "no build output found for %s"
	CRED_404 = "no credential found for %s %s"
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