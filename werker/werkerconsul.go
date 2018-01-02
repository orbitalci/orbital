package werker

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"fmt"
)

var (
	buildPath 	   = "ci/builds/%s"
	buildDonePath  = buildPath + "/done" //  %s is hash
	buildRegister  = buildPath +"/werker_ip"
)


func SetBuildDone(consulete *consul.Consulet, gitHash string) error {
	// todo: add byte of 0/1.. have ot use binary library though and idk how to use that yet
	// and not motivated enough to do it right now
	err := consulete.AddKeyValue(fmt.Sprintf(buildDonePath, gitHash), []byte("true"))
	if err != nil {
		 return err
	}
	return nil
}

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

func Register(consulete *consul.Consulet, gitHash string, ip string) error {
	err := consulete.AddKeyValue(fmt.Sprintf(buildRegister, gitHash), []byte(ip))
	return err
}

func Delete(consulete *consul.Consulet, gitHash string) error {
	err := consulete.RemoveValues(fmt.Sprintf(buildPath, gitHash))
	// for now, leaving in build done
	err = SetBuildDone(consulete, gitHash)
	return err
}