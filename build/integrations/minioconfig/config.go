package minioconfig

import (
	"encoding/json"

	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
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
	configEncoded := common.BitzToBase64(bitz)
	m.mConf = configEncoded
	return configEncoded, err
}

func (m *minioConf) MakeBashable(encoded string) []string {
	return []string{"/bin/sh", "-c", "mkdir -p ~/.mc && echo \"${MCONF}\" | base64 -d > ~/.mc/config.json"}
}

func (m *minioConf) IsRelevant(wc *pb.BuildConfig) bool {
	return common.BuildScriptsContainString(wc, "mc")
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
