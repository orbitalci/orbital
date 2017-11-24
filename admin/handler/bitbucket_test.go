package handler

import (
	"testing"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/util"
	"github.com/golang/protobuf/jsonpb"
	"os"
)

func TestBitbucket_FindWebhooks(t *testing.T) {
	config := &models.AdminConfig{}
	bb := Bitbucket{}
	bb.SetCallbackURL("marianne-callback-url")
	bb.SetMeUp(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
		Marshaler: &jsonpb.Marshaler{},
	})

	results := bb.FindWebhooks("test-find-webhooks")
	if len(results) != 2 {
		t.Error(util.GenericStrFormatErrors("callback urls that need to be created", 2, len(results)))
	}
	if results[0] != "rp" {
		t.Error(util.GenericStrFormatErrors("webhook type that needs creation", "rp", results[0]))
	}
	if results[1] != "pr" {
		t.Error(util.GenericStrFormatErrors("2nd webhook type that needs creation", "pr", results[1]))
	}
}

type MockHttpClient struct {
	Unmarshaler *jsonpb.Unmarshaler
	Marshaler *jsonpb.Marshaler
}

func (mhc MockHttpClient) GetUrl(url string, unmarshalObj proto.Message) error {
	switch url {
	case "test-find-webhooks":
		webhooks, _ := os.Open("testing/GetWebhooksResp.json")
		_ = mhc.Unmarshaler.Unmarshal(webhooks , unmarshalObj)
	}
	return nil
}

func (mhc MockHttpClient) GetUrlRawData(url string) ([]byte, error) {
	return []byte{}, nil
}

func (mhc MockHttpClient) PostUrl(url string, body string, unmarshalObj proto.Message) error {
	return nil
}