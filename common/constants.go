package common

import (
	"fmt"
	"os"
	"strings"
	"sync"
)
var once sync.Once
var prefix string
const BuildFileName = "ocelot.yml"

var BitbucketEvents = []string{"repo:push", "pullrequest:approved", "pullrequest:updated"}

func getPrefix() string {
	once.Do(func(){
		prefix = os.Getenv("PATH_PREFIX")
		if prefix == "" {
			prefix = ""
		}
		prefix = prefix + "/"
	})
	return prefix
}


var (
	OcyConfigBase = getPrefix() +  "config/ocelot"
	StorageType =  OcyConfigBase + "/storagetype"

	PostgresCredLoc = OcyConfigBase + "/postgres"
	PostgresDatabaseName = PostgresCredLoc + "/db"
	PostgresLocation = PostgresCredLoc + "/location"
	PostgresPort = PostgresCredLoc + "/port"
	PostgresUsername = PostgresCredLoc + "/username"
	PostgresPasswordLoc = "secret/" + PostgresCredLoc
	PostgresPasswordKey = "clientsecret"

	FilesystemConfigLoc =  OcyConfigBase + "/filesystem"
	FilesystemDir = FilesystemConfigLoc + "/savedirec"
	ConfigPath = "creds"
)


const (
	buildBase		= "ci/builds/"
	buildIdOnly     = buildBase + "%s" // werkerId
	buildPath 	    = buildBase + "%s/%s" // werkerId, hash
	dockerUuidKey   = "docker_uuid"
	buildDockerUuid = buildPath + "/" + dockerUuidKey
	summaryId       = "build_id"
	buildSummaryId  = buildPath + "/" + summaryId
	currentStage    = "current_stage"
	bldCurrentStage = buildPath + "/" + currentStage
	startTime       = "start_time"
	bldStartTime    = buildPath + "/" + startTime

	werkerBuildBase = "ci/werker_build_map/"
	werkerBuildMap  = werkerBuildBase + "%s" // %s is hash

	werkerLocBase   = "ci/werker_location/"
	werkerLocation  = werkerLocBase + "%s" // %s is werker id
	werkerIp        = werkerLocation + "/werker_ip"
	werkerGrpc      = werkerLocation + "/werker_grpc_port"
	werkerWs	    = werkerLocation + "/werker_ws_port"
)

func MakeBuildPath(werkerId string, gitHash string) string {
	return getPrefix() + fmt.Sprintf(buildPath, werkerId, gitHash)
}

func MakeBuildWerkerIdPath(werkerId string) string {
	return getPrefix() + fmt.Sprintf(buildIdOnly, werkerId)
}

func MakeBuildSummaryIdPath(werkerId string, gitHash string) string {
	return getPrefix() + fmt.Sprintf(buildSummaryId, werkerId, gitHash)
}

func MakeBuildStagePath(werkerId string, gitHash string) string {
	return getPrefix() + fmt.Sprintf(bldCurrentStage, werkerId, gitHash)
}

func MakeBuildStartpath(werkerId string, gitHash string) string {
	return getPrefix() + fmt.Sprintf(bldStartTime, werkerId, gitHash)
}

func MakeDockerUuidPath(werkerId string, gitHash string) string {
	return getPrefix() + fmt.Sprintf(buildDockerUuid, werkerId, gitHash)
}

func MakeBuildMapPath(gitHash string) string {
	return getPrefix() + fmt.Sprintf(werkerBuildMap, gitHash)
}

func MakeWerkerLocPath(werkerId string) string {
	return getPrefix() + fmt.Sprintf(werkerLocation, werkerId)
}

func MakeWerkerIpPath(werkerId string) string {
	return getPrefix() + fmt.Sprintf(werkerIp, werkerId)
}

func MakeWerkerGrpcPath(werkerId string) string {
	return getPrefix() + fmt.Sprintf(werkerGrpc, werkerId)
}

func MakeWerkerWsPath(werkerId string) string {
	return getPrefix() + fmt.Sprintf(werkerWs, werkerId)
}



// parseGenericBuildPath will return the werkerId and hash out of a key related to the build path
// must be fully qualified key path, not prefix
// ie: ci/builds/<werkerId>/<hash>/docker_uuid
func parseGenericBuildPath(buildPath string) (werkerId string, hash string, key string) {
	var shift int
	if getPrefix() != "" {
		shift = 1
	}
	split := strings.Split(buildPath, "/")
	werkerId = split[2+shift]
	hash = split[3+shift]
	key = split[4+shift]
	return
}


// parseBuildMapPath will return the git hash of the werkerBuildMap key
// 	ie: ci/werker_build_map/<hash>
func parseBuildMapPath(path string) (hash string) {
	split := strings.Split(path, "/")
	return split[len(split) - 1]
}


//parseWerkerLocPath will return the werkerId from a werkerLocation path configured in Consul
// ie: ci/werker_location/<werkerId>/werker_ip
// must be fully qualified key path, not prefix
func parseWerkerLocPath(path string) (werkerId string) {
	var shift int
	if getPrefix() != "" {
		shift = 1
	}
	split := strings.Split(path, "/")
	werkerId = split[2+shift]
	return
}