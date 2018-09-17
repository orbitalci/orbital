package admin

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	"github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *guideOcelotServer) BuildRuntime(ctx context.Context, bq *pb.BuildQuery) (*pb.Builds, error) {
	start := startRequest()
	defer finishRequest(start)
	if bq.Hash == "" && bq.BuildId == 0 {
		return nil, status.Error(codes.InvalidArgument, "either hash or build id is required")
	}
	buildRtInfo := make(map[string]*pb.BuildRuntimeInfo)
	var err error

	if len(bq.Hash) > 0 {
		//find matching hashes in consul by git hash
		buildRtInfo, err = build.GetBuildRuntime(g.RemoteConfig.GetConsul(), bq.Hash)
		if err != nil {
			if _, ok := err.(*build.ErrBuildDone); !ok {
				log.IncludeErrField(err).Error("could not get build runtime")
				return nil, status.Error(codes.Internal, "could not get build runtime, err: "+err.Error())
			} else {
				//we set error back to nil so that we can continue with the rest of the logic here
				err = nil
			}
		}

		//add matching hashes in db if exists and add acctname/repo to ones found in consul
		dbResults, err := g.Storage.RetrieveHashStartsWith(bq.Hash)

		if err != nil {
			return &pb.Builds{
				Builds: buildRtInfo,
			}, handleStorageError(err)
		}

		for _, bild := range dbResults {
			if _, ok := buildRtInfo[bild.Hash]; !ok {
				buildRtInfo[bild.Hash] = &pb.BuildRuntimeInfo{
					Hash: bild.Hash,
					// if a result was found in the database but not in GetBuildRuntime, the build is done
					Done: true,
				}
			}
			buildRtInfo[bild.Hash].AcctName = bild.Account
			buildRtInfo[bild.Hash].RepoName = bild.Repo
		}
	}
	//fixme: this is no longer valid to assume that just because the buildId is passed that the build is done. builds are added to the db from the _start_ of the build.
	//if a valid build id passed, go ask db for entries
	if bq.BuildId > 0 {
		buildSum, err := g.Storage.RetrieveSumByBuildId(bq.BuildId)
		if err != nil {
			return &pb.Builds{
				Builds: buildRtInfo,
			}, handleStorageError(err)
		}

		buildRtInfo[buildSum.Hash] = &pb.BuildRuntimeInfo{
			Hash:     buildSum.Hash,
			Done:     true,
			AcctName: buildSum.Account,
			RepoName: buildSum.Repo,
		}
	}

	builds := &pb.Builds{
		Builds: buildRtInfo,
	}
	return builds, err
}

// scanLog will create a scanner out of the buildOutput byte data and send it over the GuideOcelot logs stream.
//   will return a grpc error if something goes wrong
func scanLog(out models.BuildOutput, stream pb.GuideOcelot_LogsServer, storageType string) error {
	scanner := bufio.NewScanner(bytes.NewReader(out.Output))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		resp := &pb.LineResponse{OutputLine: scanner.Text()}
		stream.Send(resp)
	}
	if err := scanner.Err(); err != nil {
		log.IncludeErrField(err).Error("error encountered scanning from " + storageType)
		return status.Error(codes.Internal, fmt.Sprintf("Error was encountered while sending data from %s. \nError: %s", storageType, err.Error()))
	}
	return nil
}

// Logs will stream logs from storage. If the build is not complete, an InvalidArgument gRPC error will be returned
//   If the BuildQuery's BuildId is > 0, then logs will be retrieved from storage via the buildId. If this is not the case,
//   then the latest log entry from the hash will be retrieved and streamed.
func (g *guideOcelotServer) Logs(bq *pb.BuildQuery, stream pb.GuideOcelot_LogsServer) error {
	start := startRequest()
	defer finishRequest(start)
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
		return scanLog(out, stream, g.Storage.StorageType())
	}

	if !build.CheckIfBuildDone(g.RemoteConfig.GetConsul(), g.Storage, bq.Hash) {
		errmsg := "build is not finished, use BuildRuntime method and stream from the werker registered"
		stream.Send(&pb.LineResponse{OutputLine: errmsg})
		return status.Error(codes.InvalidArgument, errmsg)
	} else {
		out, err = g.Storage.RetrieveLastOutByHash(bq.Hash)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Unable to retrieve from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
		return scanLog(out, stream, g.Storage.StorageType())
	}
}

func (g *guideOcelotServer) FindWerker(ctx context.Context, br *pb.BuildReq) (*pb.BuildRuntimeInfo, error) {
	start := startRequest()
	defer finishRequest(start)
	if len(br.Hash) > 0 {
		//find matching hashes in consul by git hash
		buildRtInfo, err := build.GetBuildRuntime(g.RemoteConfig.GetConsul(), br.Hash)
		if err != nil {
			if _, ok := err.(*build.ErrBuildDone); !ok {
				return nil, status.Errorf(codes.Internal, "could not get build runtime, err: %s", err.Error())
			}
			return nil, status.Error(codes.InvalidArgument, "werker not found for request as it has already finished ")
		}

		if len(buildRtInfo) == 0 || len(buildRtInfo) > 1 {
			return nil, status.Error(codes.InvalidArgument, "ONE and ONE ONLY match should be found for your hash")
		}

		for _, v := range buildRtInfo {
			return v, nil
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "Please pass a hash")
	}
	return nil, nil
}
