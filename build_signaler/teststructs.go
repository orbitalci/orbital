package build_signaler

import (
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/consul/api"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/deserialize"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/vault"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
)

var Buildfile = []byte(`image: golang:1.10.2-alpine3.7
buildTool: go
env: 
  - "BUILD_DIR=/go/src/bitbucket.org/level11consulting/"
branches:
  - ALL
stages:
  - name: install consul for testing
    script: 
      - apk update 
      - apk add unzip
      - cd /go/bin 
      - wget https://releases.hashicorp.com/consul/1.1.0/consul_1.1.0_linux_amd64.zip
      - echo "unzipping"
      - unzip consul_1.1.0_linux_amd64.zip
      - echo "Done"
  - name: configure git
    script:
      - git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"
  - name: make stoopid dep thing
    script:
      - mkdir -p $BUILD_DIR
      - cp -r $WORKSPACE $BUILD_DIR/go-til
  - name: install dep & ensure dependencies
    script:
      - cd $BUILD_DIR/go-til
      - go get -u github.com/golang/dep/...
      - dep ensure -v
  - name: test
    script:
      - cd $BUILD_DIR
      - go test ./...

`)

var BuildFileMasterOnly = []byte(`image: golang:1.10.2-alpine3.7
buildTool: go
env: 
  - "BUILD_DIR=/go/src/bitbucket.org/level11consulting/"
branches:
  - master
stages:
  - name: install consul for testing
    script: 
      - apk update 
      - apk add unzip
      - cd /go/bin 
      - wget https://releases.hashicorp.com/consul/1.1.0/consul_1.1.0_linux_amd64.zip
      - echo "unzipping"
      - unzip consul_1.1.0_linux_amd64.zip
      - echo "Done"
  - name: configure git
    script:
      - git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"
  - name: make stoopid dep thing
    script:
      - mkdir -p $BUILD_DIR
      - cp -r $WORKSPACE $BUILD_DIR/go-til
  - name: install dep & ensure dependencies
    script:
      - cd $BUILD_DIR/go-til
      - go get -u github.com/golang/dep/...
      - dep ensure -v
  - name: test
    script:
      - cd $BUILD_DIR
      - go test ./...

`)

func GetFakeSignaler(t *testing.T, inConsul bool) *Signaler {
	cred := &config.RemoteConfig{Consul: &TestConsul{keyFound: inConsul}, Vault: &TestVault{}}
	dese := deserialize.New()
	valid := &build.OcelotValidator{}
	store := &TestStorage{}
	produ := &TestSingleProducer{Done: make(chan int, 1)}
	return NewSignaler(cred, dese, produ, valid, store)
}

type TestSingleProducer struct {
	Message proto.Message
	Topic   string
	Done    chan int
}

func (tp *TestSingleProducer) WriteProto(message proto.Message, topicName string) error {
	tp.Message = message
	tp.Topic = topicName
	close(tp.Done)
	return nil
}

type TestVault struct {
	vault.Vaulty
}

func (tv *TestVault) CreateThrowawayToken() (string, error) {
	return "token", nil
}

type TestConsul struct {
	consul.Consuletty
	keyFound bool
}

func (tc *TestConsul) GetKeyValue(string) (*api.KVPair, error) {
	if tc.keyFound {
		return &api.KVPair{}, nil
	}
	return nil, nil
}

// can take one build
type TestStorage struct {
	storage.OcelotStorage
	summary *pb.BuildSummary
	stages  []*models.StageResult
}

func (ts *TestStorage) AddSumStart(hash, account, repo, branch string, by pb.SignaledBy, credId int64) (int64, error) {
	ts.summary = &pb.BuildSummary{Hash: hash, Account: account, Repo: repo, Branch: branch, BuildId: 12}
	return 12, nil
}

func (ts *TestStorage) SetQueueTime(id int64) error {
	ts.summary.QueueTime = &timestamp.Timestamp{Seconds: 0, Nanos: 0}
	return nil
}

func (ts *TestStorage) StoreFailedValidation(id int64) error {
	ts.summary.Failed = true
	return nil
}

func (ts *TestStorage) AddStageDetail(result *models.StageResult) error {
	ts.stages = append(ts.stages, result)
	return nil
}

type DummyVcsHandler struct {
	Fail         bool
	Filecontents []byte
	ChangedFiles []string
	ReturnCommit *pb.Commit
	CommitNotFound bool
	models.VCSHandler
	NotFound bool
}
//fixme: need to add handler.GetChangedFiles and handler.GetCommit

func (d *DummyVcsHandler) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	if d.Fail {
		return nil, errors.New("failing")
	}
	if d.NotFound {
		return nil, ocenet.FileNotFound
	}
	return d.Filecontents, nil
}

func (d *DummyVcsHandler) GetChangedFiles(acctRepo, latesthash, earliestHash string) ([]string, error) {
	return d.ChangedFiles, nil
}

func (d *DummyVcsHandler) GetCommit(acctRepo, hash string) (*pb.Commit, error) {
	if d.CommitNotFound {
		return nil, errors.New("not found")
	}
	return d.ReturnCommit, nil
}

// set all flags to their default nil value
func (d *DummyVcsHandler) Reset() {
	d.Fail = false
	d.Filecontents = []byte{}
	d.ChangedFiles = []string{}
	d.ReturnCommit = nil
	d.CommitNotFound = false
	d.NotFound = false
}
