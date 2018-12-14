package consul

import (
	"fmt"
	"strings"
)

const (
	buildBase       = "ci/builds/"
	buildIdOnly     = buildBase + "%s"    // werkerId
	buildPath       = buildBase + "%s/%s" // werkerId, hash
	DockerUuidKey   = "docker_uuid"
	BuildDockerUuid = buildPath + "/" + DockerUuidKey
	SummaryId       = "build_id"
	buildSummaryId  = buildPath + "/" + SummaryId
	CurrentStage    = "current_stage"
	bldCurrentStage = buildPath + "/" + CurrentStage
	StartTime       = "start_time"
	bldStartTime    = buildPath + "/" + StartTime

	werkerBuildBase = "ci/werker_build_map/"
	WerkerBuildMap  = werkerBuildBase + "%s" // %s is hash

	werkerLocBase  = "ci/werker_location/"
	werkerLocation = werkerLocBase + "%s" // %s is werker id
	WerkerIp       = werkerLocation + "/werker_ip"
	WerkerGrpc     = werkerLocation + "/werker_grpc_port"
	WerkerWs       = werkerLocation + "/werker_ws_port"
	WerkerTags     = werkerLocation + "/tags"
)

func MakeBuildPath(werkerId string, gitHash string) string {
	return GetPrefix() + fmt.Sprintf(buildPath, werkerId, gitHash)
}

func MakeBuildWerkerIdPath(werkerId string) string {
	return GetPrefix() + fmt.Sprintf(buildIdOnly, werkerId)
}

func MakeBuildSummaryIdPath(werkerId string, gitHash string) string {
	return GetPrefix() + fmt.Sprintf(buildSummaryId, werkerId, gitHash)
}

func MakeBuildStagePath(werkerId string, gitHash string) string {
	return GetPrefix() + fmt.Sprintf(bldCurrentStage, werkerId, gitHash)
}

func MakeBuildStartpath(werkerId string, gitHash string) string {
	return GetPrefix() + fmt.Sprintf(bldStartTime, werkerId, gitHash)
}

func MakeDockerUuidPath(werkerId string, gitHash string) string {
	return GetPrefix() + fmt.Sprintf(BuildDockerUuid, werkerId, gitHash)
}

func MakeBuildMapPath(gitHash string) string {
	return GetPrefix() + fmt.Sprintf(WerkerBuildMap, gitHash)
}

func MakeWerkerLocPath(werkerId string) string {
	return GetPrefix() + fmt.Sprintf(werkerLocation, werkerId)
}

func MakeWerkerIpPath(werkerId string) string {
	return GetPrefix() + fmt.Sprintf(WerkerIp, werkerId)
}

func MakeWerkerGrpcPath(werkerId string) string {
	return GetPrefix() + fmt.Sprintf(WerkerGrpc, werkerId)
}

func MakeWerkerWsPath(werkerId string) string {
	return GetPrefix() + fmt.Sprintf(WerkerWs, werkerId)
}

func MakeWerkerTagsPath(werkerId string) string {
	return GetPrefix() + fmt.Sprintf(WerkerTags, werkerId)
}

// ParseGenericBuildPath will return the werkerId and hash out of a key related to the build path
// must be fully qualified key path, not prefix
// ie: ci/builds/<werkerId>/<hash>/docker_uuid
func ParseGenericBuildPath(buildPath string) (werkerId string, hash string, key string) {
	var shift int
	if GetPrefix() != "" {
		shift = 1
	}
	split := strings.Split(buildPath, "/")
	werkerId = split[2+shift]
	hash = split[3+shift]
	key = split[4+shift]
	return
}

// ParseBuildMapPath will return the git hash of the WerkerBuildMap key
// 	ie: ci/werker_build_map/<hash>
func ParseBuildMapPath(path string) (hash string) {
	split := strings.Split(path, "/")
	return split[len(split)-1]
}

//ParseWerkerLocPath will return the werkerId from a werkerLocation path configured in Consul
// ie: ci/werker_location/<werkerId>/werker_ip
// must be fully qualified key path, not prefix
func ParseWerkerLocPath(path string) (werkerId string) {
	var shift int
	if GetPrefix() != "" {
		shift = 1
	}
	split := strings.Split(path, "/")
	werkerId = split[2+shift]
	return
}
