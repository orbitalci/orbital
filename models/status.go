package models

import (
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/shankj3/ocelot/models/pb"
)

//ParseStagesByBuildId will combine the buildsummary + stages to a single object called "Status"
func ParseStagesByBuildId(buildSum *pb.BuildSummary, stageResults []StageResult) *pb.Status {
	var parsedStages []*pb.StageStatus
	for _, result := range stageResults {
		stageDupe := &pb.StageStatus{
			StageStatus:   result.Stage,
			Error:         result.Error,
			Status:        int32(result.Status),
			Messages:      result.Messages,
			StartTime:     &timestamp.Timestamp{Seconds: result.StartTime.UTC().Unix()},
			StageDuration: result.StageDuration,
		}
		parsedStages = append(parsedStages, stageDupe)
	}

	hashStatus := &pb.Status{
		BuildSum: buildSum,
		Stages: parsedStages,
	}

	return hashStatus
}
