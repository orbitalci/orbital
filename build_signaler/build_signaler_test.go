package build_signaler

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

var buildfile = []byte(`image: golang:1.10.2-alpine3.7
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
      - echo "done"
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

var expectedBuildConf = &pb.BuildConfig{
	BuildTool: "go",
	Image: "golang:1.10.2-alpine3.7",
	Env: []string{"BUILD_DIR=/go/src/bitbucket.org/level11consulting/"},
	Branches: []string{"ALL"},
	Stages: []*pb.Stage{
		{
			Name: "install consul for testing",
			Script: []string{"apk update", "apk add unzip", "cd /go/bin", "wget https://releases.hashicorp.com/consul/1.1.0/consul_1.1.0_linux_amd64.zip", "echo \"unzipping\"", "unzip consul_1.1.0_linux_amd64.zip", "echo \"done\""},
		},
		{
			Name: "configure git",
			Script: []string{`git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"`},
		},
		{
			Name: "make stoopid dep thing",
			Script: []string{"mkdir -p $BUILD_DIR", "cp -r $WORKSPACE $BUILD_DIR/go-til"},
		},
		{
			Name: "install dep & ensure dependencies",
			Script: []string{"cd $BUILD_DIR/go-til", "go get -u github.com/golang/dep/...", "dep ensure -v"},
		},
		{
			Name: "test",
			Script: []string{"cd $BUILD_DIR", "go test ./..."},
		},
	},
}

func TestCheckForBuildFile(t *testing.T) {
	dese := deserialize.New()
	conf, err := CheckForBuildFile(buildfile, dese)
	if err != nil {
		t.Error("err")
	}
	if diff := deep.Equal(conf, expectedBuildConf); diff != nil {
		t.Error(diff)
	}
}

func TestGetConfig(t *testing.T) {
	handler := &dummyVcsHandler{fail: false, filecontents:buildfile}
	conf, err := GetConfig("jessi/shank", "12345", deserialize.New(), handler)
	if err != nil {
		t.Error("should not return an error, everything should be fine.")
	}
	if diff := deep.Equal(conf, expectedBuildConf); diff != nil {
		t.Error(diff)
	}
	handler = &dummyVcsHandler{fail: true}
	_, err = GetConfig("jessi/shank", "12345", deserialize.New(), handler)
	if err == nil {
		t.Error("vcs handler returned a failure, that should be bubbled up")
	}
	_, err = GetConfig("jessi/shank", "12345", deserialize.New(), nil)
	if err == nil {
		t.Error("vcs handler is nil, should return an error")
	}
}

type dummyVcsHandler struct {
	fail bool
	filecontents []byte
	models.VCSHandler
}

func (d *dummyVcsHandler) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	if d.fail {
		return nil, errors.New("failing")
	}
	return d.filecontents, nil
}

