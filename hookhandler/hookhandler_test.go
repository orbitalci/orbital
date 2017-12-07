package main

import (
	"testing"
	pb "github.com/shankj3/ocelot/protos/out"
	"fmt"
	"bitbucket.org/level11consulting/go-til/test"
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
	if buildOnlyPipe == nil {
		t.Error(fmt.Sprintf("expected %v to NOT be null", buildOnlyPipe))
	}

	if buildOnlyPipe.GlobalEnv == nil {
		t.Error(fmt.Sprintf("expected %v to NOT be null", buildOnlyPipe.GlobalEnv))
	}

	if buildOnlyPipe.GlobalEnv["DEBUG"] != "1" {
		t.Error(test.GenericStrFormatErrors("global env DEBUG", "1", buildOnlyPipe.GlobalEnv["DEBUG"]))
	}

	if buildOnlyPipe.Steps == nil {
		t.Error(fmt.Sprintf("expected %v to NOT be null", buildOnlyPipe.Steps))
	}

	if len(buildOnlyPipe.Steps) != 1 {
		t.Error(test.GenericStrFormatErrors("pipeline steps", 1, len(buildOnlyPipe.Steps)))
	}

	if buildOnlyPipe.Steps["marianne"] == nil {
		t.Error(fmt.Sprintf("expected %v to NOT be null", buildOnlyPipe.Steps["marianne"] ))
	}

	if buildOnlyPipe.Steps["marianne"].Image != "TODO PARSE THIS AND PUSH TO ARTIFACT REPO" {
		t.Error(test.GenericStrFormatErrors("job image", "TODO PARSE THIS AND PUSH TO ARTIFACT REPO", buildOnlyPipe.Steps["marianne"].Image))
	}

	if buildOnlyPipe.Steps["marianne"].Command != "sh -a && cp -r . .." {
		t.Error(test.GenericStrFormatErrors("pipeline command", "sh -a && cp -r . ..", buildOnlyPipe.Steps["marianne"].Command))
	}

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
	if buildOnlyPipe == nil {
		t.Error(fmt.Sprintf("expected %v to NOT be null", buildOnlyPipe))
	}

	if len(buildOnlyPipe.GlobalEnv) > 0 {
		t.Error(fmt.Sprintf("expected %v to be empty", buildOnlyPipe.GlobalEnv))
	}

	if buildOnlyPipe.Steps == nil {
		t.Error(fmt.Sprintf("expected %v to NOT be null", buildOnlyPipe.Steps))
	}

	if len(buildOnlyPipe.Steps) != 1 {
		t.Error(test.GenericStrFormatErrors("pipeline steps", 1, len(buildOnlyPipe.Steps)))
	}

	if buildOnlyPipe.Steps["marianne"] == nil {
		t.Error(fmt.Sprintf("expected %v to NOT be null", buildOnlyPipe.Steps["marianne"] ))
	}

	if buildOnlyPipe.Steps["marianne"].Image != "wowanimage" {
		t.Error(test.GenericStrFormatErrors("job image", "wowanimage", buildOnlyPipe.Steps["marianne"].Image))
	}

	if buildOnlyPipe.Steps["marianne"].Command != "first && sh -a && cp -r . .." {
		t.Error(test.GenericStrFormatErrors("pipeline command", "first && sh -a && cp -r . ..", buildOnlyPipe.Steps["marianne"].Command))
	}

	if buildOnlyPipe.Steps["marianne"].Env["DEBUG"] != "1" {
		t.Error(test.GenericStrFormatErrors("build stage env DEBUG", "1", buildOnlyPipe.Steps["marianne"].Env["DEBUG"] ))
	}

	if buildOnlyPipe.Steps["marianne"].Env["SOMETHING"] != "ELSE" {
		t.Error(test.GenericStrFormatErrors("build stage env SOMETHING", "ELSE", buildOnlyPipe.Steps["marianne"].Env["SOMETHING"] ))
	}
}


