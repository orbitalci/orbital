package admin

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


func (g *guideOcelotServer) BuildRuntime(ctx context.Context, bq *pb.BuildQuery) (*pb.Builds, error) {
	buildRtInfo := make(map[string]*pb.BuildRuntimeInfo)
	var err error

	if len(bq.Hash) > 0 {
		//find matching hashes in consul by git hash
		buildRtInfo, err = build.GetBuildRuntime(g.RemoteConfig.GetConsul(), bq.Hash)
		if err != nil {
			if _, ok := err.(*build.ErrBuildDone); !ok {
				log.IncludeErrField(err)
				return nil, status.Error(codes.Internal, "could not get build runtime, err: " + err.Error())
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

		for _, build := range dbResults {
			if _, ok := buildRtInfo[build.Hash]; !ok {
				buildRtInfo[build.Hash] = &pb.BuildRuntimeInfo{
					Hash: build.Hash,
					// if a result was found in the database but not in GetBuildRuntime, the build is done
					Done: true,
				}
			}
			buildRtInfo[build.Hash].AcctName = build.Account
			buildRtInfo[build.Hash].RepoName = build.Repo
		}
	}

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



func (g *guideOcelotServer) Logs(bq *pb.BuildQuery, stream pb.GuideOcelot_LogsServer) error {
	if !build.CheckIfBuildDone(g.RemoteConfig.GetConsul(), g.Storage, bq.Hash) {
		errmsg :=  "build is not finished, use BuildRuntime method and stream from the werker registered"
		stream.Send(&pb.LineResponse{OutputLine: errmsg})
		return status.Error(codes.NotFound,  errmsg)
	} else {
		var out models.BuildOutput
		var err error
		if bq.BuildId != 0 {
			out, err = g.Storage.RetrieveOut(bq.BuildId)
		} else {
			out, err = g.Storage.RetrieveLastOutByHash(bq.Hash)
		}
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("Unable to retrieve from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
		scanner := bufio.NewScanner(bytes.NewReader(out.Output))
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			resp := &pb.LineResponse{OutputLine: scanner.Text()}
			stream.Send(resp)
		}
		if err := scanner.Err(); err != nil {
			log.IncludeErrField(err).Error("error encountered scanning from " + g.Storage.StorageType())
			return status.Error(codes.Internal, fmt.Sprintf("Error was encountered while sending data from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
	}
	return nil
}


func (g *guideOcelotServer) FindWerker(ctx context.Context, br *pb.BuildReq) (*pb.BuildRuntimeInfo, error) {
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