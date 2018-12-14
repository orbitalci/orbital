/*
  minioconfig is an implementation of the StringIntegrator interface

	minioconfig's methods will use all minio repo credentials and generate a mc-compatible config.json. builds can then use the mc minio client
	to access the object store without any additional configuration
*/
package minioconfig

import (
	"encoding/json"

	"github.com/level11consulting/orbitalci/build/helpers/buildscript/search"
	"github.com/level11consulting/orbitalci/build/helpers/serde"
	"github.com/level11consulting/orbitalci/build/integrations"
	"github.com/level11consulting/orbitalci/models/pb"
)

func Create() integrations.StringIntegrator {
	return &minioConf{}
}

type minioConf struct {
	mConf string
}

func (m *minioConf) String() string {
	return "minio log in"
}

func (m *minioConf) SubType() pb.SubCredType {
	return pb.SubCredType_MINIO
}

func (m *minioConf) GenerateIntegrationString(credz []pb.OcyCredder) (string, error) {
	bitz, err := rcToMinioConf(credz)
	if err != nil {
		return "", err
	}
	configEncoded := serde.BitzToBase64(bitz)
	m.mConf = configEncoded
	return configEncoded, err
}

func (m *minioConf) MakeBashable(encoded string) []string {
	return []string{"/bin/sh", "-c", "mkdir -p ~/.mc && echo \"${MCONF}\" | base64 -d > ~/.mc/config.json"}
}

func (m *minioConf) IsRelevant(wc *pb.BuildConfig) bool {
	return search.BuildScriptsContainString(wc, "mc")
}

func (m *minioConf) GetEnv() []string {
	return []string{"MCONF=" + m.mConf}
}

// rcToMinioConf will iterate over all credentials and build out a minio config file, keying the hosts with the cred's identifier.
// <mc>AccessKey == <repoCred>Username
// <mc>SecretKey == <repoCred>Password
// <mc>Url		 == <repoCred>RepoUrl.
func rcToMinioConf(creds []pb.OcyCredder) ([]byte, error) {
	conf := getMinioObj()
	for _, credi := range creds {
		mCred, ok := credi.(*pb.RepoCreds)
		if !ok {
			continue
		}
		mConfEntry := &minioConfigEntry{
			Url:       mCred.RepoUrl,
			AccessKey: mCred.Username,
			SecretKey: mCred.Password,
			Api:       defaultApi,
			Lookup:    defaultLookup,
		}
		conf.Hosts[mCred.Identifier] = mConfEntry
	}
	return json.Marshal(conf)
}
