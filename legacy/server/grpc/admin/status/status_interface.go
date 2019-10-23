package status

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/orbitalci/client/runtime"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/models/pb"
	metrics "github.com/level11consulting/orbitalci/server/metrics/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/log"
	storage_error "github.com/level11consulting/orbitalci/storage/error"
	"github.com/level11consulting/orbitalci/server/config"
	"github.com/level11consulting/orbitalci/server/grpc/admin/sendstream"
	"github.com/level11consulting/orbitalci/server/grpc/admin/scanoutput"
	"github.com/level11consulting/orbitalci/storage"
)

type StatusInterface interface {
	CheckConn(context.Context, *empty.Empty) (*empty.Empty, error)
	GetStatus(context.Context, *pb.StatusQuery) (*pb.Status, error)
	Logs(*pb.BuildQuery, pb.GuideOcelot_LogsServer) error
	LastFewSummaries(context.Context, *pb.RepoAccount) (*pb.Summaries, error)
}

type StatusInterfaceAPI struct {
	StatusInterface
	RemoteConfig   config.CVRemoteConfig
	Storage        storage.OcelotStorage
}

// for checking if the server is reachable
func (g *StatusInterfaceAPI) CheckConn(ctx context.Context, msg *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

// FIXME: Not using the protobuf message will let us move error handling code somewhere else. Possibly even squash this into 3 cases?
//StatusByHash will retrieve you the status (build summary + stages) of a partial git hash
// We handle 4 cases to identify a build id, then we use the BUILD_FOUND goto label to initialize `result`.
// Providing a build id
// Providing a partial git hash (Partial from the front of the hash)
// Providing an account and repo name, which returns the first summary in the list (perhaps this is the latest?)
// Providing a "partial repo", which ends up deriving an account, and then continues to reimplement the 3rd case
//
// When we initialize `result`, we are getting "stage detail" with the build id. (Possibly all of the output?)
// We then parse the output.
// We do some other mysterious check for if a build is in consul, and set some flag on `result`
// Then we do an implicit bare return. Which is a thing in go, which could be easier to read if this weren't such a long function.
func (g *StatusInterfaceAPI) GetStatus(ctx context.Context, query *pb.StatusQuery) (result *pb.Status, err error) {
	var buildSum *pb.BuildSummary
	switch {
	case query.BuildId != 0:
		buildSum, err = g.Storage.RetrieveSumByBuildId(query.BuildId)
		if err != nil {
			return nil, storage_error.HandleStorageError(err)
		}
		goto BUILD_FOUND
	case len(query.Hash) > 0:
		partialHash := query.Hash
		buildSum, err = g.Storage.RetrieveLatestSum(partialHash)
		if err != nil {
			return nil, storage_error.HandleStorageError(err)
		}
		goto BUILD_FOUND

	case len(query.AcctName) > 0 && len(query.RepoName) > 0:
		buildSums, err := g.Storage.RetrieveLastFewSums(query.RepoName, query.AcctName, 1)
		if err != nil {
			return nil, storage_error.HandleStorageError(err)
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
	case len(query.PartialRepo) > 0:
		buildSums, err := g.Storage.RetrieveAcctRepo(strings.TrimSpace(query.PartialRepo))
		if err != nil {
			return nil, storage_error.HandleStorageError(err)
		}

		if len(buildSums) == 1 {
			buildSumz, err := g.Storage.RetrieveLastFewSums(buildSums[0].Repo, buildSums[0].Account, 1)
			if err != nil {
				return nil, storage_error.HandleStorageError(err)
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
			log.IncludeErrField(uhOh).Error()
			return nil, status.Error(codes.InvalidArgument, uhOh.Error())
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "either hash is required, acctName and repoName is required, or partialRepo is required")
	}
BUILD_FOUND:
	stageResults, err := g.Storage.RetrieveStageDetail(buildSum.BuildId)
	if err != nil {
		return nil, storage_error.HandleStorageError(err)
	}
	result = models.ParseStagesByBuildId(buildSum, stageResults)
	// idk if htis is necessary anymore
	inConsul, err := runtime.CheckBuildInConsul(g.RemoteConfig.GetConsul(), buildSum.Hash)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "An error occurred checking build status in consul. Cannot retrieve status at this time.\n\n"+err.Error())
	}
	result.IsInConsul = inConsul
	return
}

// Logs will stream logs from storage. If the build is not complete, an InvalidArgument gRPC error will be returned
//   If the BuildQuery's BuildId is > 0, then logs will be retrieved from storage via the buildId. If this is not the case,
//   then the latest log entry from the hash will be retrieved and streamed.
func (g *StatusInterfaceAPI) Logs(bq *pb.BuildQuery, stream pb.GuideOcelot_LogsServer) error {
	start := metrics.StartRequest()
	defer metrics.FinishRequest(start)
	if bq.Hash == "" && bq.BuildId == 0 {
		return status.Error(codes.InvalidArgument, "must request with either a hash or a buildId")
	}
	var out models.BuildOutput
	var err error
	if bq.BuildId != 0 {
		out, err = g.Storage.RetrieveOut(bq.BuildId)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Unable to retrive from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
		return scanoutput.ScanLog(out, stream, g.Storage.StorageType(), bq.Strip)
	}

	if !runtime.CheckIfBuildDone(g.RemoteConfig.GetConsul(), g.Storage, bq.Hash) {

		errmsg := "build is not finished, use BuildRuntime method and stream from the werker registered"
		sendstream.SendStream(stream, errmsg)
		return status.Error(codes.InvalidArgument, errmsg)
	} else {
		out, err = g.Storage.RetrieveLastOutByHash(bq.Hash)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Unable to retrieve from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
		return scanoutput.ScanLog(out, stream, g.Storage.StorageType(), bq.Strip)
	}
}


func (g *StatusInterfaceAPI) LastFewSummaries(ctx context.Context, repoAct *pb.RepoAccount) (*pb.Summaries, error) {
	if repoAct.Repo == "" || repoAct.Account == "" || repoAct.Limit == 0 {
		return nil, status.Error(codes.InvalidArgument, "repo, account, and limit are required fields")
	}
	var summaries = &pb.Summaries{}
	modelz, err := g.Storage.RetrieveLastFewSums(repoAct.Repo, repoAct.Account, repoAct.Limit)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, storage_error.HandleStorageError(err)
	}
	log.Log().Debug("successfully retrieved last few summaries")
	if len(modelz) == 0 {
		return nil, status.Error(codes.NotFound, "no entries found")
	}
	for _, model := range modelz {
		summaries.Sums = append(summaries.Sums, model)
	}
	//fmt.Println(summaries)
	return summaries, nil

}
