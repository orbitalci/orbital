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
	//"encoding/binary"
	//"errors"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
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

type HashRuntime struct {
	DockerUuid   string
	BuildId 	 int64
	CurrentStage string
	Hash         string
	StageStart   time.Time
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

func UnRegister(consulete *consul.Consulet, werkerId string) error {
	err := consulete.RemoveValues(MakeWerkerLocPath(werkerId))
	return err
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

func RegisterBuildSummaryId(consulete *consul.Consulet, werkerId string, gitHash string, buildId int64) error {
	str := fmt.Sprintf("%d", buildId)
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("buildId", buildId).Info("registering build")
	err := consulete.AddKeyValue(MakeBuildSummaryIdPath(werkerId, gitHash), []byte(str))
	return err
}


func RegisterBuildStage(consulete *consul.Consulet, werkerId string, gitHash string, buildStage string) error {
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("buildStage", buildStage).Info("registering build")
	err := consulete.AddKeyValue(MakeBuildStagePath(werkerId, gitHash), []byte(buildStage))
	return err
}

func RegisterStageStartTime(consulete *consul.Consulet, werkerId string, gitHash string, start time.Time) error {
	str := fmt.Sprintf("%d", start.Unix()) // todo: figure out a better way to do this conversion using bit shifting or something because i know this isnt the "right" way to do it
	err := consulete.AddKeyValue(MakeBuildStartpath(werkerId, gitHash), []byte(str))
	return err
}


func GetDockerUuidsByWerkerId(consulete *consul.Consulet, werkerId string) (uuids []string, err error) {
	pairs, err := consulete.GetKeyValues(MakeBuildWerkerIdPath(werkerId))
	if err != nil {
		return
	}
	for _, pair := range pairs {
		if strings.Contains(pair.Key, dockerUuidKey) {
			uuids = append(uuids, string(pair.Value))
		}
	}
	return
}

func GetHashRuntimesByWerker(consulete *consul.Consulet, werkerId string) (hrts map[string]*HashRuntime, err error){
	pairs, err := consulete.GetKeyValues(MakeBuildWerkerIdPath(werkerId))
	hrts = make(map[string]*HashRuntime)
	if err != nil { return }
	for _, pair := range pairs {
		_, hash, key := parseGenericBuildPath(pair.Key)
		_, ok := hrts[hash]
		if !ok {
			hrts[hash] = &HashRuntime{
				Hash: hash,
			}
		}
		switch key {
		case dockerUuidKey:
			hrts[hash].DockerUuid = string(pair.Value)
		case summaryId:
			var id int64
			id, err = convertArrayToInt(pair.Value)
			hrts[hash].BuildId = id
		case currentStage:
			hrts[hash].CurrentStage = string(pair.Value)
		case startTime:
			var unix int64
			unix, err = convertArrayToInt(pair.Value)
			startTime := time.Unix(unix, 0)
			hrts[hash].StageStart = startTime
		}
	}
	return
}

func convertArrayToInt(array []byte) (int64, error) {
	integ, err := strconv.Atoi(string(array))
	return int64(integ), err
}

//func main() {
//	var n int64
//	b := [8]byte{1, 2}
//	buf := bytes.NewBuffer(&b)
//	binary.Read(buf, binary.LittleEndian, &n)
//	fmt.Println(n, b)
//}


func GetWerkerActiveBuilds(consulete *consul.Consulet, werkerId string) (hashes []string, err error) {
	// todo: allow for a different separator? will we ever be using a different one? probably not, but technically you can...
	keys, err := consulete.GetKeys(MakeBuildWerkerIdPath(werkerId))
	if err != nil {
		return
	}
	s := map[string]bool{}
	for _, key := range keys {
		//fmt.Println(key)
		//ind := strings.LastIndex(key, "/")
		//hashInd := strings.LastIndex(key[:ind+1], "/")
		//hashes = append(hashes, key[hashInd+1:])
		_, hash, _ := parseGenericBuildPath(key)
		_, ok := s[hash]
		if !ok {
			hashes = append(hashes, hash)
			s[hash] = true
		}
	}
	return
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
		ocelog.IncludeErrField(err).Error("couldn't get kv error!")
		return
	}
	if kv == nil {
		ocelog.Log().Error("THIS PAIR SHOULD NOT BE NIL! path: " + pairPath)
		return
	}
	ocelog.Log().WithField("gitHash", gitHash).Info("WERKERID IS: ", string(kv.Value))
	if err = consulete.RemoveValues(MakeBuildPath(string(kv.Value), gitHash)); err != nil {
		return
	}
	err = consulete.RemoveValue(pairPath)
	return err
}


func CheckBuildInConsul(consulete *consul.Consulet, hash string) (exists bool, err error) {
	pairPath := MakeBuildMapPath(hash)
	kv, err := consulete.GetKeyValue(pairPath)
	if err != nil {
		return
	}
	if kv != nil {
		exists = true
	}
	return
}