package commandhelper

import (
	"bitbucket.org/level11consulting/ocelot/old/admin/models"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
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
	Debuggit(cmd.GetUI(), "checking connection of " + cmd.GetConfig().AdminLocation)
	_, err := cmd.GetClient().CheckConn(ctx, &empty.Empty{})
	if err != nil {
		cmd.GetUI().Error(err.Error())
		cmd.GetUI().Error(fmt.Sprintf("could not connect to server at %s", cmd.GetConfig().AdminLocation))
	}
	return err
}

type GuideOcelotCmdStdin interface {
	GuideOcelotCmd
	GetDataFromUiAsk(ui cli.Ui) (interface{}, error)
}
