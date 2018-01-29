package buildruntime

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"fmt"
	"strings"
	"bitbucket.org/level11consulting/ocelot/admin/models"
)

var (
	buildPath 	   = "ci/builds/%s"
	buildDonePath  = buildPath + "/done" //  %s is hash
	buildRegister  = buildPath +"/werker_ip"
	buildGrpcPort  = buildPath + "/werker_grpc_port"
	buildWsPort    = buildPath + "/werker_ws_port"
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

func GetBuildRuntime(consulete *consul.Consulet, gitHash string) (map[string]*models.BuildRuntimeInfo, error) {
	path := fmt.Sprintf(buildPath, gitHash)
	pairs, err := consulete.GetKeyValues(path)
	if err != nil {
		return nil, err
	}
	rt := make(map[string]*models.BuildRuntimeInfo)
	if len(pairs) == 0 {
		//rt.Done = true
		return rt, &ErrBuildDone{"no build found in consul"}
	}

	for _, pair := range pairs {
		key := pair.Key[strings.LastIndex(pair.Key, "/") + 1:]
		keySub := pair.Key[:strings.LastIndex(pair.Key, "/")]
		gitHash := keySub[strings.LastIndex(keySub, "/") + 1:]
		_, ok := rt[gitHash]
		if !ok {
			rt[gitHash] = &models.BuildRuntimeInfo{
				Hash: gitHash,
			}
		}

		switch key {
		case "done":
			rt[gitHash].Done = true
		case "werker_ip":
			rt[gitHash].Ip = string(pair.Value)
		case "werker_grpc_port":
			rt[gitHash].GrpcPort = string(pair.Value)
		case "werker_ws_port":
			// don't use this right now
		}
	}
	return rt, nil
}

// SetBuildDone adds the flag `ci/builds/<gitHash>/done` to consul
func SetBuildDone(consulete *consul.Consulet, gitHash string) error {
	// todo: add byte of 0/1.. have ot use binary library though and idk how to use that yet
	// and not motivated enough to do it right now
	err := consulete.AddKeyValue(fmt.Sprintf(buildDonePath, gitHash), []byte("true"))
	if err != nil {
		 return err
	}
	return nil
}

// CheckIfBuildDone will check in consul to make sure there is nothing in runtime configuration anymore,
// then it will makes sure it can find it in storage
func CheckIfBuildDone(consulete *consul.Consulet, summary storage.BuildSum, gitHash string) bool {
	kv, err := consulete.GetKeyValue(fmt.Sprintf(buildRegister, gitHash))
	fmt.Println("KV!", kv)
	if err != nil {
		// log here what the err is, etc
		fmt.Println(err)
		return false
	}
	if kv != nil {
		return false
	} else {
		// look in storage if not found in consul
		_, err := summary.RetrieveLatestSum(gitHash)
		if err != nil {
			if _, ok := err.(*storage.ErrNotFound); !ok {
				// log here what the err is, etc
				fmt.Println(err)
				 return false
			} else { return true }
		}
		return true
	}
}

// Register will add all the appropriate build details that the admin needs to contact the werker for stream info
// will add:
// ci/builds/<gitHash>/ + werker_ip        = ip
// 		'' 			    + werker_grpc_port = grpcPort
// 		''				+ werker_ws_port   = wsPort
func Register(consulete *consul.Consulet, gitHash string, ip string, grpcPort string, wsPort string) (err error) {
	// TODO: add in postgres so we can get rid of consul done path at the end (see line 63)
	if err = consulete.RemoveValue(fmt.Sprintf(buildDonePath, gitHash)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(fmt.Sprintf(buildRegister, gitHash), []byte(ip)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(fmt.Sprintf(buildGrpcPort, gitHash), []byte(grpcPort)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(fmt.Sprintf(buildWsPort, gitHash), []byte(wsPort)); err != nil {
		return
	}
	return
}

// Delete will remove everything related to that werker's build of the gitHash out of consul
func Delete(consulete *consul.Consulet, gitHash string) error {
	err := consulete.RemoveValues(fmt.Sprintf(buildPath, gitHash))
	// for now, leaving in build done
	//err = SetBuildDone(consulete, gitHash)
	return err
}