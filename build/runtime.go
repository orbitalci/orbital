package build

import (
	"strconv"
	"strings"
	"time"

	"github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/level11consulting/ocelot/common"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
)

type HashRuntime struct {
	DockerUuid   string
	BuildId      int64
	CurrentStage string
	Hash         string
	StageStart   time.Time
}

//GetBuildRuntime will return BuildRuntimeInfo about matching partial git hashes.
//It does this by first asking consul for state of builds, and then db storage
func GetBuildRuntime(consulete consul.Consuletty, gitHash string) (map[string]*pb.BuildRuntimeInfo, error) {
	mapPath := common.MakeBuildMapPath(gitHash)
	pairs, err := consulete.GetKeyValues(mapPath)
	if err != nil {
		return nil, err
	}
	rt := make(map[string]*pb.BuildRuntimeInfo)
	if len(pairs) == 0 {
		return rt, &ErrBuildDone{"no build found in consul"}
	}
	for _, pair := range pairs {
		fullHash := common.ParseBuildMapPath(pair.Key)
		werkerId := string(pair.Value)
		locPairs, err := consulete.GetKeyValues(common.MakeWerkerLocPath(werkerId))
		if err != nil {
			// todo: wrap these errors so we know where they came from / at what action
			return nil, err
		}
		for _, pair := range locPairs {
			key := pair.Key[strings.LastIndex(pair.Key, "/")+1:]
			_, ok := rt[fullHash]
			if !ok {
				rt[fullHash] = &pb.BuildRuntimeInfo{
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
				rt[fullHash].WsPort = string(pair.Value)
			}
		}
	}

	return rt, nil
}

// CheckIfBuildDone will check in consul to make sure there is nothing in runtime configuration anymore,
// then it will makes sure it can find it in storage
func CheckIfBuildDone(consulete consul.Consuletty, summary storage.BuildSum, gitHash string) bool {
	kv, err := consulete.GetKeyValue(common.MakeBuildMapPath(gitHash))
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
			} else {
				return true
			}
		}
		return true
	}
}

func GetWerkerActiveBuilds(consulete consul.Consuletty, werkerId string) (hashes []string, err error) {
	// todo: allow for a different separator? will we ever be using a different one? probably not, but technically you can...
	keys, err := consulete.GetKeys(common.MakeBuildWerkerIdPath(werkerId))
	if err != nil {
		return
	}
	s := map[string]bool{}
	for _, key := range keys {
		//fmt.Println(key)
		//ind := strings.LastIndex(key, "/")
		//hashInd := strings.LastIndex(key[:ind+1], "/")
		//hashes = append(hashes, key[hashInd+1:])
		_, hash, _ := common.ParseGenericBuildPath(key)
		_, ok := s[hash]
		if !ok {
			hashes = append(hashes, hash)
			s[hash] = true
		}
	}
	return
}

func CheckBuildInConsul(consulete consul.Consuletty, hash string) (exists bool, err error) {
	pairPath := common.MakeBuildMapPath(hash)
	kv, err := consulete.GetKeyValue(pairPath)
	if err != nil {
		return
	}
	if kv != nil {
		exists = true
	}
	return
}

func GetHashRuntimesByWerker(consulete consul.Consuletty, werkerId string) (hrts map[string]*HashRuntime, err error) {
	pairs, err := consulete.GetKeyValues(common.MakeBuildWerkerIdPath(werkerId))
	hrts = make(map[string]*HashRuntime)
	if err != nil {
		return
	}
	for _, pair := range pairs {
		_, hash, key := common.ParseGenericBuildPath(pair.Key)
		_, ok := hrts[hash]
		if !ok {
			hrts[hash] = &HashRuntime{
				Hash: hash,
			}
		}
		switch key {
		case common.DockerUuidKey:
			hrts[hash].DockerUuid = string(pair.Value)
		case common.SummaryId:
			var id int64
			id, err = convertArrayToInt(pair.Value)
			hrts[hash].BuildId = id
		case common.CurrentStage:
			hrts[hash].CurrentStage = string(pair.Value)
		case common.StartTime:
			var unix int64
			unix, err = convertArrayToInt(pair.Value)
			startTime := time.Unix(unix, 0)
			hrts[hash].StageStart = startTime
		}
	}
	return
}

func GetDockerUuidsByWerkerId(consulete consul.Consuletty, werkerId string) (uuids []string, err error) {
	pairs, err := consulete.GetKeyValues(common.MakeBuildWerkerIdPath(werkerId))
	if err != nil {
		return
	}
	for _, pair := range pairs {
		if strings.Contains(pair.Key, common.DockerUuidKey) {
			uuids = append(uuids, string(pair.Value))
		}
	}
	return
}

func convertArrayToInt(array []byte) (int64, error) {
	integ, err := strconv.Atoi(string(array))
	return int64(integ), err
}

type ErrBuildDone struct {
	msg string
}

func (e *ErrBuildDone) Error() string {
	return e.msg
}
