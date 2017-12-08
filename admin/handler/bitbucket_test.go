package handler

import (
	"testing"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/util"
	"github.com/golang/protobuf/jsonpb"
	"os"
)


func TestBitbucket_FindWebhooksExists(t *testing.T) {
	config := &models.Credentials{}
	bb := Bitbucket{}
	bb.SetCallbackURL("webhook-exists-url")
	bb.SetMeUp(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	})

	results := bb.FindWebhooks("webhooks-exists")
	if !results {
		t.Error(util.GenericStrFormatErrors("webhook exists", true, results))
	}
}

func TestBitbucket_FindWebhooksEmpty(t *testing.T) {
	config := &models.Credentials{}
	bb := Bitbucket{}
	bb.SetCallbackURL("marianne-empty-callback-url")
	bb.SetMeUp(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	})

	results := bb.FindWebhooks("empty-webhooks")
	if results {
		t.Error(util.GenericStrFormatErrors("no webhook yet", false, results))
	}
}

type MockHttpClient struct {
	Unmarshaler *jsonpb.Unmarshaler
}

func (mhc MockHttpClient) GetUrl(url string, unmarshalObj proto.Message) error {
	switch url {
	case "empty-webhooks":
		webhooks, _ := os.Open("test-fixtures/EmptyWebhooksResp.json")
		_ = mhc.Unmarshaler.Unmarshal(webhooks , unmarshalObj)
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