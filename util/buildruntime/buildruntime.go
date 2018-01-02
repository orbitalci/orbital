package buildruntime

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"fmt"
)

var (
	buildPath 	   = "ci/builds/%s"
	buildDonePath  = buildPath + "/done" //  %s is hash
	buildRegister  = buildPath +"/werker_ip"
	buildGrpcPort  = buildPath + "/werker_grpc_port"
	buildWsPort    = buildPath + "/werker_ws_port"
)

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

// CheckIfBuildDone will do a check in consul for the `done` flag
// todo: should also look in db if not in consul
func CheckIfBuildDone(consulete *consul.Consulet, gitHash string) bool {
	kv, err := consulete.GetKeyValue(fmt.Sprintf(buildDonePath, gitHash))
	if err != nil {
		// idk what we should be doing if the error is not nil, maybe panic? hope that never happens?
		return false
	}
	if kv != nil {
		return true
	}
	return false
}

// Register will add all the appropriate build details that the admin needs to contact the werker for stream info
// will add:
// ci/builds/<gitHash>/ + werker_ip        = ip
// 		'' 			    + werker_grpc_port = grpcPort
// 		''				+ werker_ws_port   = wsPort
func Register(consulete *consul.Consulet, gitHash string, ip string, grpcPort string, wsPort string) (err error) {
	if err := consulete.AddKeyValue(fmt.Sprintf(buildRegister, gitHash), []byte(ip)); err != nil {
		return
	}
	if err := consulete.AddKeyValue(fmt.Sprintf(buildGrpcPort, gitHash), []byte(grpcPort)); err != nil {
		return
	}
	if err := consulete.AddKeyValue(fmt.Sprintf(buildWsPort, gitHash), []byte(wsPort)); err != nil {
		return
	}
	return
}

// Delete will remove everything related to that werker's build of the gitHash out of consul
// **HOWEVER... RIGHT NOW...** it will leave in the `done` flag (until we include a postgres db)
// This should be called after a build has completed and everything has been stored.
func Delete(consulete *consul.Consulet, gitHash string) error {
	err := consulete.RemoveValues(fmt.Sprintf(buildPath, gitHash))
	// for now, leaving in build done
	err = SetBuildDone(consulete, gitHash)
	return err
}