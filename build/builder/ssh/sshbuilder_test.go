package ssh

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

func TestSSH_Setup(t *testing.T) {
	ssher, ctx, cancel, tarRm, cleaner := SetupSSHBuilder(t, 2222, "3833")
	defer tarRm(t)
	defer cleaner()
	defer cancel()
	logout := make(chan []byte, 1000)
	defer close(logout)
	dockerId := make(chan string, 1)
	result, _ := ssher.Setup(ctx, logout, dockerId, &pb.WerkerTask{CheckoutHash:"TESTHASHAYYY"}, nil, "")
	expectedMsgs := []string{"attempting to establish ssh connection...", "successfully established ssh connection " + models.CHECKMARK, "Set up via SSH " + models.CHECKMARK}
	if diff := deep.Equal(result.Messages, expectedMsgs); diff != nil {
		t.Error(diff)
	}
}
