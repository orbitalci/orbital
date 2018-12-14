package builddb

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/level11consulting/orbitalci/models/pb"
	"github.com/shankj3/go-til/deserialize"
)

var expectedBuildConf = &pb.BuildConfig{
	BuildTool: "go",
	Image:     "golang:1.10.2-alpine3.7",
	Env:       []string{"BUILD_DIR=/go/src/bitbucket.org/level11consulting/"},
	Branches:  []string{"ALL"},
	Stages: []*pb.Stage{
		{
			Name:   "install consul for testing",
			Script: []string{"apk update", "apk add unzip", "cd /go/bin", "wget https://releases.hashicorp.com/consul/1.1.0/consul_1.1.0_linux_amd64.zip", "echo \"unzipping\"", "unzip consul_1.1.0_linux_amd64.zip", "echo \"Done\""},
		},
		{
			Name:   "configure git",
			Script: []string{`git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"`},
		},
		{
			Name:   "make stoopid dep thing",
			Script: []string{"mkdir -p $BUILD_DIR", "cp -r $WORKSPACE $BUILD_DIR/go-til"},
		},
		{
			Name:   "install dep & ensure dependencies",
			Script: []string{"cd $BUILD_DIR/go-til", "go get -u github.com/golang/dep/...", "dep ensure -v"},
		},
		{
			Name:   "test",
			Script: []string{"cd $BUILD_DIR", "go test ./..."},
		},
	},
}

func TestCheckForBuildFile(t *testing.T) {
	dese := deserialize.New()
	conf, err := CheckForBuildFile(Buildfile, dese)
	if err != nil {
		t.Error("err")
	}
	if diff := deep.Equal(conf, expectedBuildConf); diff != nil {
		t.Error(diff)
	}
}

func TestGetConfig(t *testing.T) {
	handler := &DummyVcsHandler{Fail: false, Filecontents: Buildfile}
	conf, err := GetConfig("jessi/shank", "12345", deserialize.New(), handler)
	if err != nil {
		t.Error("should not return an error, everything should be fine.")
	}
	if diff := deep.Equal(conf, expectedBuildConf); diff != nil {
		t.Error(diff)
	}
	handler = &DummyVcsHandler{Fail: true}
	_, err = GetConfig("jessi/shank", "12345", deserialize.New(), handler)
	if err == nil {
		t.Error("vcs handler returned a failure, that should be bubbled up")
	}
	_, err = GetConfig("jessi/shank", "12345", deserialize.New(), nil)
	if err == nil {
		t.Error("vcs handler is nil, should return an error")
	}
}
