package main

import (
	"testing"
	"github.com/shankj3/ocelot/util/test"
	pb "github.com/shankj3/ocelot/protos/out"
)

func TestHookhandler_WerkBuildOnly(t *testing.T) {
	testEnv := make(map[string]string)
	testEnv["DEBUG"] = "1"

	buildOnly := pb.BuildConfig{
		Env: testEnv,
		Packages: []string{"fakepkg1"},
		Build: &pb.Stage{
			Env: testEnv,
			Script: []string{"sh -a", "cp -r . .."},
		},
	}

	buildOnlyPipe, _ := werk(buildOnly, "marianne")
	test.AssertNotNull(t, buildOnlyPipe)
	test.AssertNotNull(t, buildOnlyPipe.GlobalEnv)
	test.GenericAssertEqs(t, "1", buildOnlyPipe.GlobalEnv["DEBUG"])
	test.AssertNotNull(t, buildOnlyPipe.Steps)
	test.GenericAssertEqs(t, 1, len(buildOnlyPipe.Steps))
	test.AssertNotNull(t, buildOnlyPipe.Steps["marianne"])
	test.GenericAssertEqs(t, "TODO PARSE THIS AND PUSH TO ARTIFACT REPO", buildOnlyPipe.Steps["marianne"].Image)
	test.GenericAssertEqs(t, "sh -a && cp -r . ..", buildOnlyPipe.Steps["marianne"].Command)

	if buildOnlyPipe.Steps["marianne"].Env["DEBUG"] != testEnv["DEBUG"] {
		t.Error(test.GenericStrFormatErrors("build stage env DEBUG", "1", buildOnlyPipe.Steps["marianne"].Env["DEBUG"] ))
	}
}

func TestHookhandler_WerkBeforeAndBuild(t *testing.T) {
	testEnv := make(map[string]string)
	testEnv["DEBUG"] = "1"
	testEnv["SOMETHING"] = "LAST"

	testEnv2 := make(map[string]string)
	testEnv2["SOMETHING"] = "ELSE"

	buildOnly := pb.BuildConfig{
		Image: "wowanimage",
		Before: &pb.Stage{
			Env: testEnv,
			Script: []string{"first"},
		},
		Build: &pb.Stage{
			Env: testEnv2,
			Script: []string{"sh -a", "cp -r . .."},
		},
	}

	buildOnlyPipe, _ := werk(buildOnly, "marianne")
	test.AssertNotNull(t, buildOnlyPipe)
	test.AssertNotNull(t, buildOnlyPipe.GlobalEnv)
	test.AssertNull(t, buildOnlyPipe.GlobalEnv["DEBUG"])
	test.AssertNotNull(t, buildOnlyPipe.Steps)
	test.GenericAssertEqs(t, 1, len(buildOnlyPipe.Steps))
	test.AssertNotNull(t, buildOnlyPipe.Steps["marianne"])
	test.GenericAssertEqs(t, "wowanimage", buildOnlyPipe.Steps["marianne"].Image)
	test.GenericAssertEqs(t, "first && sh -a && cp -r . ..", buildOnlyPipe.Steps["marianne"].Command)

	if buildOnlyPipe.Steps["marianne"].Env["DEBUG"] != "1" {
		t.Error(test.GenericStrFormatErrors("build stage env DEBUG", "1", buildOnlyPipe.Steps["marianne"].Env["DEBUG"] ))
	}

	if buildOnlyPipe.Steps["marianne"].Env["SOMETHING"] != "ELSE" {
		t.Error(test.GenericStrFormatErrors("build stage env SOMETHING", "ELSE", buildOnlyPipe.Steps["marianne"].Env["SOMETHING"] ))
	}
}


