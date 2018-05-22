package build_signaler

import (
	"testing"

	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models/pb"
)

func Test_validateBuild(t *testing.T) {
	buildConf := &pb.BuildConfig{
		Image: "busybox:latest",
		BuildTool: "w/e",
		Branches: []string{"rc_*"},
		Stages: []*pb.Stage{
			{Name: "hi", Script: []string{"echo sup"}},
		},
	}
	validator := build.GetOcelotValidator()
	err := validateBuild(buildConf, "rc_1234", validator)
	if err != nil {
		t.Error("validation should pass, error is : " + err.Error())
	}
	err = validateBuild(buildConf, "r1_1234", validator)
	if err == nil {
		t.Error("should not pass validation")
	}
	errMsg := "branch r1_1234 does not match any branches listed: [rc_*]"
	if err.Error() != errMsg {
		t.Error(test.StrFormatErrors("err msg", errMsg, err.Error()))
	}
}
