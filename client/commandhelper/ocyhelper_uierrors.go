package commandhelper

import(
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (oh *OcyHelper) LastFewSummariesErr(erro error, cmd GuideOcelotCmd) {
	unknownErr := `Unable to retrieve the last few summaries for Account %s and Repository %s.

Please see the maintainers about the following error: %s`
	notFound := `There have been no build summaries for Account %s and Repository %s.
Please see the documentation about how to add a repository in Ocelot.`
	err, ok := status.FromError(erro)
	if !ok {
		oh.WriteUi(cmd.GetUI().Error, fmt.Sprintf(unknownErr, oh.Account, oh.Repo, erro.Error()))
		return
	}
	if err.Code() == codes.NotFound {
		oh.WriteUi(cmd.GetUI().Warn, fmt.Sprintf(notFound, oh.Account, oh.Repo))
	} else {
		oh.WriteUi(cmd.GetUI().Error, fmt.Sprintf(unknownErr, oh.Account, oh.Repo, err.Message()))
	}
}