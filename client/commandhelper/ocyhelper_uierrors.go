package commandhelper

import (
	"fmt"

	"github.com/mitchellh/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (oh *OcyHelper) LastFewSummariesErr(erro error, ui cli.Ui) {
	unknownErr := `Unable to retrieve the last few summaries for Account %s and Repository %s.

Please see the maintainers about the following error: %s`
	notFound := `There have been no build summaries for Account %s and Repository %s.
Please see the documentation about how to add a repository in Ocelot.`
	err, ok := status.FromError(erro)
	if !ok {
		oh.WriteUi(ui.Error, fmt.Sprintf(unknownErr, oh.Account, oh.Repo, erro.Error()))
		return
	}
	if err.Code() == codes.NotFound {
		oh.WriteUi(ui.Warn, fmt.Sprintf(notFound, oh.Account, oh.Repo))
	} else {
		oh.WriteUi(ui.Error, fmt.Sprintf(unknownErr, oh.Account, oh.Repo, err.Message()))
	}
}

type DontOverwrite struct {
}

func (d *DontOverwrite) Error() string {
	return "chose not to overwrite"
}
