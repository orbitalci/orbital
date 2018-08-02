package minioconfig

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
)

var expected = []byte(`{
  "version": "9",
  "hosts": {
      "mypurtyminio": {
        "url": "https://minio.go",
        "accessKey": ";lsakdfjak",
        "secretKey": "ksdjlfaklsdfj",
        "api": "s3v4",
        "lookup": "auto"
      },
	  "theminiobaaaaby": {
        "url": "https://minio.gorgeous",
        "accessKey": "thisisaaccesskey",
        "secretKey": "thisisasecretkey",
        "api": "s3v4",
        "lookup": "auto"
      }
  }
}`)

func TestMinioConf_GenerateIntegrationString(t *testing.T) {
	minio := Create()
	credz := []pb.OcyCredder{
		&pb.RepoCreds{
			SubType: pb.SubCredType_MINIO,
			RepoUrl: "https://minio.go",
			Password: "ksdjlfaklsdfj",
			Username: ";lsakdfjak",
			Identifier: "mypurtyminio",
		},
		&pb.RepoCreds{
			SubType: pb.SubCredType_MINIO,
			RepoUrl: "https://minio.gorgeous",
			Password: "thisisasecretkey",
			Username: "thisisaaccesskey",
			Identifier: "theminiobaaaaby",
		},
	}
	expectedStruct := getMinioObj()
	err := json.Unmarshal(expected, expectedStruct)
	if err != nil {
		t.Error(err)
	}
	generatedString, err := minio.GenerateIntegrationString(credz)
	if err != nil {
		t.Error(err)
	}

	bitz, err  := common.Base64ToBitz(string(generatedString))
	if err != nil {
		t.Error(err)
	}
	liveStruct := getMinioObj()
	err = json.Unmarshal(bitz, liveStruct)
	if err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(expectedStruct, liveStruct); diff != nil {
		t.Error(diff)
	}
}

func TestMinioConf_IsRelevant(t *testing.T) {
	wc := &pb.BuildConfig{
		Stages: []*pb.Stage{
			{Script: []string{"echo 'hi' && mc create"}},
		},
	}
	minio := Create()
	if !minio.IsRelevant(wc) {
		t.Error("should be relevant, mc is in script")
	}
}

func TestMinioConf_MakeBashable(t *testing.T) {
	m := Create()
	generate := m.MakeBashable("not applicable")
	expect := []string{"/bin/sh", "-c", "mkdir -p ~/.mc && echo \"${MCONF}\" | base64 -d > ~/.mc/config.json"}
	if diff := deep.Equal(expect, generate); diff != nil {
		t.Error(diff)
	}
}

func TestMinioConf_SubType(t *testing.T) {
	m := Create()
	if m.SubType() != pb.SubCredType_MINIO {
		t.Error("how did this get set incorrectly?? wrong subtype")
	}
}

func TestMinioConf_String(t *testing.T) {
	m := Create()
	if m.String() != "minio log in" {
		t.Error("wrong minio integration string, should be 'minio log in'")
	}
}