package bitbucket

import (
	"net/http"
	"os"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
)

func TestBitbucket_FindWebhooksExists(t *testing.T) {
	config := &pb.VCSCreds{}
	bb := GetBitbucketHandler(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	})
	bb.SetCallbackURL("webhook-exists-url")
	results := bb.FindWebhooks("webhooks-exists")
	if !results {
		t.Error(test.GenericStrFormatErrors("webhook exists", true, results))
	}
}

func TestBitbucket_FindWebhooksEmpty(t *testing.T) {
	config := &pb.VCSCreds{}
	bb := GetBitbucketHandler(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	})
	bb.SetCallbackURL("marianne-empty-callback-url")
	results := bb.FindWebhooks("empty-webhooks")
	if results {
		t.Error(test.GenericStrFormatErrors("no webhook yet", false, results))
	}
}

type MockHttpClient struct {
	Unmarshaler *jsonpb.Unmarshaler
}

func (mhc MockHttpClient) GetUrl(url string, unmarshalObj proto.Message) error {
	switch url {
	case "empty-webhooks":
		webhooks, _ := os.Open("test-fixtures/EmptyWebhooksResp.json")
		_ = mhc.Unmarshaler.Unmarshal(webhooks, unmarshalObj)
	case "webhooks-exists":
		webhooks, _ := os.Open("test-fixtures/WebhookExistsResp.json")
		_ = mhc.Unmarshaler.Unmarshal(webhooks, unmarshalObj)
	}
	return nil
}

func (mhc MockHttpClient) GetUrlRawData(url string) ([]byte, error) {
	return []byte{}, nil
}

func (mhc MockHttpClient) PostUrl(url string, body string, unmarshalObj proto.Message) error {
	return nil
}

func (mhc MockHttpClient) GetUrlResponse(url string) (*http.Response, error) {
	return nil, nil
}
