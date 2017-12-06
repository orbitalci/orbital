package main

import (
	"testing"
	"github.com/shankj3/ocelot/util"
	pb "github.com/shankj3/ocelot/protos/out"
)

func TestHookhandler_Werk(t *testing.T) {
	testEnv := make(map[string]string)
	testEnv["DEBUG"] = "1"

	buildOnly := pb.BuildConfig{
		Env: testEnv,
		Build: &pb.Stage{
			Env: testEnv,
			Script: []string{"sh -a", "cp -r . .."},
		},
	}

	buildOnlyPipe, _ := werk(buildOnly, "marianne")
	util.NotNull(t, buildOnlyPipe)
	util.NotNull(t, buildOnlyPipe.GlobalEnv)
	util.NotNull(t, buildOnlyPipe.Steps)
}
