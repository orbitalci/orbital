package builder

import (
	"context"
	"os/exec"
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/build"
)

func Test_runCommandLogToChan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	logout := make(chan []byte, 1000)
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", "echo hi; sleep 2; echo againhi; sleep 1; exit 0")
	err := runCommandLogToChan(cmd, logout, build.InitStageUtil("test"))
	close(logout)
	if err != nil {
		t.Error(err)
	}
	expected := []string{"hi", "againhi"}
	var live []string
	for i := range logout {
		live = append(live, string(i))
	}
	if diff := deep.Equal(expected, live); diff != nil {
		t.Error(diff)
	}
	cancel()
}