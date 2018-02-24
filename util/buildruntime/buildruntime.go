package buildruntime
/*
How an active build would actually look in consul:
	$ # Hash = abcgithash
	$ # Werker ID = werker123
	$ consul kv get -recurse ci/werker_build_map
	ci/werker_build_map/abcgithash:werker123
	ci/werker_build_map/123githash:otherwerker
	$ consul kv get -recurse ci/werker_location/werker123
	ci/werker_location/werker123/werker_ip:10.1.1.1
	ci/werker_location/werker123/werker_grpc_port:8901
	ci/werker_location/werker_ws_port:9999
	$ consul kv get ci/werker_build_map/abcgithash
	werker123
	$ consul kv get -recurse ci/builds/werker123/abcgithash
	ci/builds/werker123/abcgithash/done:false
	ci/builds/werker123/abcgithash/docker_uuid:1233-5679-5894-1111
	$ # this would enable a lookup of all active builds under a werker
	$ consul kv get -recurse ci/builds/werker123/
	ci/builds/werker123/abcgithash/docker_uuid:1233-5679-5494-1111
	ci/builds/werker123/otheractivegithash/docker_uuid:1243-5679-5894-1111
*/

import (
	"bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"github.com/google/uuid"
	"strings"
)


type BuildRuntime struct {
	Done 	 bool
	Ip   	 string
	GrpcPort string
	WsPort   string
	Hash	string
}

type ErrBuildDone struct {
	msg string
}

func (e *ErrBuildDone) Error() string {
	return e.msg
}


//GetBuildRuntime will return BuildRuntimeInfo about matching partial git hashes.
//It does this by first asking consul for state of builds, and then db storage
func GetBuildRuntime(consulete *consul.Consulet, gitHash string) (map[string]*models.BuildRuntimeInfo, error) {
	mapPath := MakeBuildMapPath(gitHash)
	pairs, err := consulete.GetKeyValues(mapPath)
	if err != nil {
		return nil, err
	}
	rt := make(map[string]*models.BuildRuntimeInfo)
	if len(pairs) == 0 {
		return rt, &ErrBuildDone{"no build found in consul"}
	}
	for _, pair := range pairs {
		fullHash := parseBuildMapPath(pair.Key)
		werkerId := string(pair.Value)
		locPairs, err := consulete.GetKeyValues(MakeWerkerLocPath(werkerId))
		if err != nil {
			// todo: wrap these errors so we know where they came from / at what action
			return nil, err
		}
		for _, pair := range locPairs {
			key := pair.Key[strings.LastIndex(pair.Key, "/") + 1:]
			_, ok := rt[fullHash]
			if !ok {
				rt[fullHash] = &models.BuildRuntimeInfo{
					Hash: fullHash,
				}
			}

			switch key {
			case "werker_ip":
				rt[fullHash].Ip = string(pair.Value)
			case "werker_grpc_port":
				rt[fullHash].GrpcPort = string(pair.Value)
			case "werker_ws_port":
				// don't use this right now
			}
		}
	}

	return rt, nil
}
//
//// SetBuildDone adds the flag `ci/builds/<werkerId>/<gitHash>/done` to consul
//func SetBuildDone(consulete *consul.Consulet, gitHash string, werkerId string) error {
//	// todo: add byte of 0/1.. have ot use binary library though and idk how to use that yet
//	// and not motivated enough to do it right now
//	err := consulete.AddKeyValue(fmt.Sprintf(buildDonePath, werkerId, gitHash), []byte("true"))
//	if err != nil {
//		 return err
//	}
//	return nil
//}

// CheckIfBuildDone will check in consul to make sure there is nothing in runtime configuration anymore,
// then it will makes sure it can find it in storage
func CheckIfBuildDone(consulete *consul.Consulet, summary storage.BuildSum, gitHash string) bool {
	kv, err := consulete.GetKeyValue(MakeBuildMapPath(gitHash))
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return false
	}
	if kv != nil {
		return false
	} else {
		// look in storage if not found in consul
		_, err := summary.RetrieveLatestSum(gitHash)
		if err != nil {
			if _, ok := err.(*storage.ErrNotFound); !ok {
				ocelog.IncludeErrField(err).Error()
				return false
			} else { return true }
		}
		return true
	}
}

// Register will add all the appropriate build details that the admin needs to contact the werker for stream info
// will add:
// werkerLocation  = "ci/werker_location/%s" // %s is werker id
// ci/werker_location/<werkid> + werker_ip        = ip
// 		'' 			           + werker_grpc_port = grpcPort
// 		''				       + werker_ws_port   = wsPort
// returns a generated uuid for the werker
func Register(consulete *consul.Consulet, ip string, grpcPort string, wsPort string) (werkerId uuid.UUID, err error) {
	werkerId = uuid.New()
	strId := werkerId.String()
	if err = consulete.AddKeyValue(MakeWerkerIpPath(strId), []byte(ip)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(MakeWerkerGrpcPath(strId), []byte(grpcPort)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(MakeWerkerWsPath(strId), []byte(wsPort)); err != nil {
		return
	}
	return
}

func RegisterStartedBuild(consulete *consul.Consulet, werkerId string, gitHash string) error {
	if err := consulete.AddKeyValue(MakeBuildMapPath(gitHash), []byte(werkerId)); err != nil {
		return err
	}
	return nil
}

func RegisterBuild(consulete *consul.Consulet, werkerId string, gitHash string, dockerUuid string) error {
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("docker_uuid", dockerUuid).Info("registering build")
	err := consulete.AddKeyValue(MakeDockerUuidPath(werkerId, gitHash), []byte(dockerUuid))
	return err
}

// Delete will remove everything related to that werker's build of the gitHash out of consul
// will delete:
// 		ci/werker_build_map/<hash>
// 		ci/builds/<werkerId>/<hash>/*
func Delete(consulete *consul.Consulet, gitHash string) (err error) {
	//paths := &Identifiers{GitHash: gitHash}
	pairPath := MakeBuildMapPath(gitHash)
	kv, err := consulete.GetKeyValue(pairPath)
	if err != nil {
		return
	}
	if kv == nil {
		ocelog.Log().Error("THIS PAIR SHOULD NOT BE NIL! path: " + pairPath)
		return
	}
	if err = consulete.RemoveValues(MakeBuildPath(string(kv.Value), gitHash)); err != nil {
		return
	}
	err = consulete.RemoveValue(pairPath)
	// for now, leaving in build done
	//err = SetBuildDone(consulete, gitHash)
	return err
}