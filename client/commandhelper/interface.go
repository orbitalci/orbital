package commandhelper

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	models "github.com/shankj3/ocelot/models/pb"
	"google.golang.org/grpc/status"
)

type GuideOcelotCmd interface {
	GetClient() models.GuideOcelotClient
	GetUI() cli.Ui
	GetConfig() *ClientConfig
}

// CheckConnection calls the CheckConn() method of GuideOcelotClient which is validates a connection exists.
// if the connection fails, an error will be printed to the UI and the grpc err will be returned (so you can exit 1
// on the command line)
func CheckConnection(cmd GuideOcelotCmd, ctx context.Context) error {
	Debuggit(cmd.GetUI(), "checking connection of "+cmd.GetConfig().AdminLocation)
	_, err := cmd.GetClient().CheckConn(ctx, &empty.Empty{})
	if err != nil {
		if _, ok := status.FromError(err); !ok {
			cmd.GetUI().Error(err.Error())
		}
		cmd.GetUI().Error(fmt.Sprintf("Could not connect to server! Admin location is %s", cmd.GetConfig().AdminLocation))
	}
	return err
}

type GuideOcelotCmdStdin interface {
	GuideOcelotCmd
	GetDataFromUiAsk(ui cli.Ui) (interface{}, error)
}
