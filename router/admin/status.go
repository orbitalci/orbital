package admin

import (
	"context"
	"fmt"
	"strings"


	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"

)

//StatusByHash will retrieve you the status (build summary + stages) of a partial git hash
func (g *guideOcelotServer) GetStatus(ctx context.Context, query *pb.StatusQuery) (result *pb.Status, err error) {
	var buildSum models.BuildSummary
	if len(query.Hash) > 0 {
		partialHash := query.Hash
		buildSum, err = g.Storage.RetrieveLatestSum(partialHash)
		if err != nil {
			return nil, handleStorageError(err)
		}
		goto BUILD_FOUND

	}
	if len(query.AcctName) > 0 && len(query.RepoName) > 0 {
		buildSums, err := g.Storage.RetrieveLastFewSums(query.RepoName, query.AcctName, 1)
		if err != nil {
			return nil, handleStorageError(err)
		}
		if len(buildSums) == 1 {
			buildSum = buildSums[0]
			goto BUILD_FOUND
		} else if len(buildSums) == 0 {
			uhOh := errors.New(fmt.Sprintf("There are no entries that match the acctname/repo %s/%s", query.AcctName, query.RepoName))
			log.IncludeErrField(uhOh).Error()
			return nil, status.Error(codes.NotFound, uhOh.Error())
		} else {
			// todo: this is logging even when there isn't a match in the db, probably an issue with RetrieveLastFewSums not returning error if there are no rows
			uhOh := errors.New(fmt.Sprintf("there is no ONE entry that matches the acctname/repo %s/%s", query.AcctName, query.RepoName))
			log.IncludeErrField(uhOh)
			return nil, status.Error(codes.InvalidArgument, uhOh.Error())
		}
	}

	if len(query.PartialRepo) > 0 {
		buildSums, err := g.Storage.RetrieveAcctRepo(strings.TrimSpace(query.PartialRepo))
		if err != nil {
			return nil, handleStorageError(err)
		}

		if len(buildSums) == 1 {
			buildSumz, err := g.Storage.RetrieveLastFewSums(buildSums[0].Repo, buildSums[0].Account, 1)
			if err != nil {
				return nil, handleStorageError(err)
			}
			buildSum = buildSumz[0]
			goto BUILD_FOUND
		} else {
			var uhOh error
			if len(buildSums) == 0 {
				uhOh = errors.New(fmt.Sprintf("there are no repositories starting with %s", query.PartialRepo))
			} else {
				var matches []string
				for _, buildSum := range buildSums {
					matches = append(matches, buildSum.Account+"/"+buildSum.Repo)
				}
				uhOh = errors.New(fmt.Sprintf("there are %v repositories starting with %s: %s", len(buildSums), query.PartialRepo, strings.Join(matches, ",")))
			}
			log.IncludeErrField(uhOh)
			return nil, status.Error(codes.InvalidArgument, uhOh.Error())
		}
	}
	return
BUILD_FOUND:
	stageResults, err := g.Storage.RetrieveStageDetail(buildSum.BuildId)
	if err != nil {
		return nil, handleStorageError(err)
	}
	result = ParseStagesByBuildId(buildSum, stageResults)
	inConsul, err := build.CheckBuildInConsul(g.RemoteConfig.GetConsul(), buildSum.Hash)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "An error occurred checking build status in consul. Cannot retrieve status at this time.\n\n" + err.Error())
	}
	result.IsInConsul = inConsul
	return
}