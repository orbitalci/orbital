package download

import (
	"errors"
	"fmt"

	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
)

//todo: pull out check for vcsHandler == nil logic, then this can be just GetConfig()
// todo (cont): write something in remote to switch between subtypes to instantiate the correct VCSHandler implementation
//GetConfig returns the protobuf ocelot.yaml, a valid bitbucket token belonging to that repo, and possible err.
//If a VcsHandler is passed, this method will use the existing handler to retrieve the bb config. In that case,
//***IT WILL NOT RETURN A VALID TOKEN FOR YOU - ONLY BUILD CONFIG***
func GetConfig(repoFullName string, checkoutCommit string, deserializer *deserialize.Deserializer, vcsHandler models.VCSHandler) (*pb.BuildConfig, error) {
	//var bbHandler remote.VCSHandler
	//var token string

	if vcsHandler == nil {
		return nil, errors.New("vcs handler cannot be nul")
	}

	fileBytz, err := vcsHandler.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, err
	}

	conf, err := CheckForBuildFile(fileBytz, deserializer)
	return conf, err
}

//CheckForBuildFile will try to retrieve an ocelot.yaml file for a repository and return the protobuf message
func CheckForBuildFile(buildFile []byte, deserializer *deserialize.Deserializer) (*pb.BuildConfig, error) {
	conf := &pb.BuildConfig{}
	fmt.Println(string(buildFile))
	if err := deserializer.YAMLToStruct(buildFile, conf); err != nil {
		if err != ocenet.FileNotFound {
			log.IncludeErrField(err).Error("unable to get build conf")
			return conf, err
		}
		log.Log().Debugf("no ocelot yml found")
		return conf, err
	}
	return conf, nil
}
