package main

import (
	"testing"
	"github.com/shankj3/ocelot/util"
)

func TestHookhandler_ConvertStage(t *testing.T) {
	shouldBeNil := convertStageToJob(nil, "")
	if shouldBeNil != nil {
		util.GenericStrFormatErrors("empty stage", nil, shouldBeNil)
	}
}
