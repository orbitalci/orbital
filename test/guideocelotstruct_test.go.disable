package admin

import (
	"context"
	"testing"
)

func TestGuideOcelotServer_CheckConn(t *testing.T) {
	//so dumb hate code coverage sometimes
	gos := &guideOcelotServer{}
	ctx := context.Background()
	_, err := gos.CheckConn(ctx, nil)
	if err != nil {
		t.Error(err)
	}
}
