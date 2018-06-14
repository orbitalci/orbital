package build

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
)

var badConfigTests = []struct{
	name string
	badConf *pb.BuildConfig
	expectedErrMsg string
}{
	{
		name: "no image or tag",
		badConf:&pb.BuildConfig{
			//Image:     "busybox:latest",
			//MachineTag: "hi",
			BuildTool: "maven",
			Packages:  []string{},
			Branches:  []string{"ALL"},
			Env:       []string{},
			Stages:    []*pb.Stage{{Name: "one"}, {Name: "build"}},
		},
		expectedErrMsg: "uh-oh, there is no image AND no machineTag listed inside of your ocelot yaml file... one of these is required",
	},
	{
		name: "image and tag",
		badConf: &pb.BuildConfig{
			Image:     "busybox:latest",
			MachineTag: "hi",
			BuildTool: "maven",
			Packages:  []string{},
			Branches:  []string{"ALL"},
			Env:       []string{},
			Stages:    []*pb.Stage{{Name: "one"}, {Name: "build"}},
		},
		expectedErrMsg: "you cannot have both image and machineTag. they are mutually exclusive",
	},
	{
		name: "no build tool",
		badConf: &pb.BuildConfig{
			Image:     "busybox:latest",
			//BuildTool: "maven",
			Packages:  []string{},
			Branches:  []string{"ALL"},
			Env:       []string{},
			Stages:    []*pb.Stage{{Name: "one"}, {Name: "boop!"}},
		},
		expectedErrMsg: "BuildTool must be specified",
	},
	{
		name: "no stages",
		badConf: &pb.BuildConfig{
			Image:     "busybox:latest",
			BuildTool: "maven",
			Packages:  []string{},
			Branches:  []string{"ALL"},
			Env:       []string{},
			Stages:    []*pb.Stage{},
		},
		expectedErrMsg: "there must be at least one stage listed",
	},
	{
		name: "bad image",
		badConf: &pb.BuildConfig{
			Image:     "adlskfja893balkxc72",
			BuildTool: "maven",
			Packages:  []string{},
			Branches:  []string{"ALL"},
			Env:       []string{},
			Stages:    []*pb.Stage{{Name: "one"}, {Name: "boop!"}},
		},
		expectedErrMsg: `An error has occured while trying to pull for image adlskfja893balkxc72. 
Full Error is Using default tag: latest

Error response from daemon: pull access denied for adlskfja893balkxc72, repository does not exist or may require 'docker login'
. `,
	},
}

func TestOcelotValidator_ValidateConfig_short(t *testing.T) {
	valid8r := GetOcelotValidator()
	var err error
	for _, tt := range badConfigTests {
		t.Run(tt.name, func(t *testing.T) {
			err = valid8r.ValidateConfig(tt.badConf, nil)
			if err == nil {
				t.Error("validation should fail, all these configs are bad")
				return
			}
			if err.Error() != tt.expectedErrMsg {
				t.Error(test.StrFormatErrors("err msg", tt.expectedErrMsg, err.Error()))
			}
		})
	}
}

func TestOcelotValidator_ValidateConfig(t *testing.T) {
	goodconfig := &pb.BuildConfig{
		Image:     "busybox:latest",
		BuildTool: "maven",
		Packages:  []string{},
		Branches:  []string{"ALL"},
		Env:       []string{},
		Stages:    []*pb.Stage{{Name: "one"}, {Name: "build"}},
	}
	goodconfig2 := &pb.BuildConfig{
		MachineTag: "ay",
		BuildTool: "maven",
		Packages:  []string{},
		Branches:  []string{"ALL"},
		Env:       []string{},
		Stages:    []*pb.Stage{{Name: "one"}, {Name: "build"}},
	}
	valid8r := GetOcelotValidator()
	err := valid8r.ValidateConfig(goodconfig, nil)
	if err != nil {
		t.Error(err)
	}
	err = valid8r.ValidateConfig(goodconfig2, nil)
	if err != nil {
		t.Error(err)
	}
	badConfig := &pb.BuildConfig{
		Image:     "busybox:latest",
		BuildTool: "maven",
		Packages:  []string{},
		Branches:  []string{"ALL"},
		Env:       []string{},
		Stages:    []*pb.Stage{{Name: "one"}},
	}
	err = valid8r.ValidateConfig(badConfig, nil)
	if err != nil {
		t.Error("should be nil, this conf is a-ok. error is: " + err.Error())
	}
	// todo: flag this as integration test
	// privateRepo should be something that you have creds to
	privateRepo, ok := os.LookupEnv("PRIVATE_REGISTRY")
	if !ok {
		t.Log("using default privateRepo of docker.metaverse.l11.com")
		privateRepo = "docker.metaverse.l11.com"
	}
	dckrconfig, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.docker/config.json"))
	if err != nil {
		t.Log("skipping nexus pull test because couldn't get docker config")
		return
	}
	if !strings.Contains(string(dckrconfig), privateRepo) {
		t.Log("skipping nexus pull test because " + privateRepo + " is not in docker config")
		return
	}
	goodPrivateConfig := &pb.BuildConfig{
		Image:     privateRepo + "/busybox:test_do_not_delete",
		BuildTool: "maven",
		Packages:  []string{},
		Branches:  []string{"ALL"},
		Env:       []string{},
		Stages:    []*pb.Stage{{Name: "one"}, {Name: "build"}},
	}
	err = valid8r.ValidateConfig(goodPrivateConfig, nil)
	if err != nil {
		t.Log("this may have failed becasue " + privateRepo + "/busybox:test_do_not_delete is not in metaverse anymore")
		t.Error(err)
	}

}

func TestValidateBranchAgainstConf(t *testing.T) {
	ocyv := &OcelotValidator{}
	buildConf := &pb.BuildConfig{
			Image: "busybox:latest",
			BuildTool: "w/e",
			Branches: []string{"rc_.*"},
			Stages: []*pb.Stage{
				{Name: "hi", Script: []string{"echo sup"}},
			},
	}
	err := ocyv.ValidateBranchAgainstConf(buildConf, "rc_1234")
	if err != nil {
		t.Error("should be queuable, error is: " + err.Error())
	}
	err = ocyv.ValidateBranchAgainstConf(buildConf, "r1_1234")
	if err == nil {
		t.Error("should not be quueable, error is " + err.Error())
	}
	if err != nil {
		if _, ok := err.(*NotViable); !ok {
			t.Error("should be a do not queue error, instead the error is: " + err.Error())
		}
	}
}


func TestOcelotValidator_ValidateViability(t *testing.T) {
	valid8r := GetOcelotValidator()
	if err := valid8r.ValidateViability("branch", []string{"bra.*h", "banana"}, []*pb.Commit{}, false); err != nil {
		t.Error(err)
	}
	if err := valid8r.ValidateViability("branch", []string{"branc", "banna"}, []*pb.Commit{}, false); err == nil {
		t.Error("should have failed validation")
	} else if err.Error() != "branch branch not in the acceptable branches list: branc, banna" {
		t.Error(test.StrFormatErrors("error message", "branch branch not in the acceptable branches list: branc, banna", err.Error()))
	}

	expectedErr := "build will not be queued because one of [skip ci] | [ci skip] was found in the commit with hash abcd. the full commit message is its time to be skipped [ci skip]"
	if err := valid8r.ValidateViability("branch", []string{"branch"}, []*pb.Commit{{Message: "its time to be skipped [ci skip]", Hash:"abcd"}}, false); err == nil {
		t.Error("should have failed commit validation - msg has [ci skip]")
	} else if err.Error() != expectedErr {
		t.Error(test.StrFormatErrors("error message", expectedErr, err.Error()))
	}
	if err := valid8r.ValidateViability("bran", []string{"brain"}, []*pb.Commit{}, true); err != nil {
		t.Error("build was forced, should not return error")
	}
	if err := valid8r.ValidateViability("bran", []string{"bran"}, nil, false); err != nil {
		t.Error("branch valid, commits list is nil. this should pass.")
	}
}