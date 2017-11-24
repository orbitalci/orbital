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
	config := &models.AdminConfig{}
	bb := Bitbucket{}
	bb.SetCallbackURL("webhook-exists-url")
	bb.SetMeUp(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	})

	results := bb.FindWebhooks("webhooks-exists")
	if len(results) != 0 {
		t.Error(util.GenericStrFormatErrors("callback urls that need to be created", 0, len(results)))
	}
}

func TestBitbucket_FindWebhooksEmpty(t *testing.T) {
	config := &models.AdminConfig{}
	bb := Bitbucket{}
	bb.SetCallbackURL("marianne-empty-callback-url")
	bb.SetMeUp(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	})

	results := bb.FindWebhooks("empty-webhooks")
	if len(results) != 2 {
		t.Error(util.GenericStrFormatErrors("callback urls that need to be created", 2, len(results)))
	}

	var containsRp bool
	var containsPr bool

	for _, webhookType := range results {
		if webhookType == "rp" {
			containsRp = true
		}
		if webhookType == "pr" {
			containsPr = true
		}
	}
	if !containsRp {
		t.Error(util.GenericStrFormatErrors("contains rp webhook type", true, containsRp))
	}
	if !containsPr {
		t.Error(util.GenericStrFormatErrors("contains pr webhook type", true, containsPr))
	}
}

func TestBitbucket_FindWebhooks(t *testing.T) {
	config := &models.AdminConfig{}
	bb := Bitbucket{}
	bb.SetCallbackURL("marianne-callback-url")
	bb.SetMeUp(config, MockHttpClient{
		Unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	})

	results := bb.FindWebhooks("1-webhook")
	if len(results) != 1 {
		t.Error(util.GenericStrFormatErrors("callback urls that need to be created", 1, len(results)))
	}

	var containsPr bool

	for _, webhookType := range results {
		if webhookType == "pr" {
			containsPr = true
		}
	}
	if !containsPr {
		t.Error(util.GenericStrFormatErrors("contains pr webhook type", true, containsPr))
	}
}

type MockHttpClient struct {
	Unmarshaler *jsonpb.Unmarshaler
}

func (mhc MockHttpClient) GetUrl(url string, unmarshalObj proto.Message) error {
	switch url {
	case "1-webhook":
		webhooks, _ := os.Open("testing/GetWebhooksResp.json")
		_ = mhc.Unmarshaler.Unmarshal(webhooks , unmarshalObj)
	case "empty-webhooks":
		webhooks, _ := os.Open("testing/EmptyWebhooksResp.json")
		_ = mhc.Unmarshaler.Unmarshal(webhooks , unmarshalObj)
	case "webhooks-exists":
		webhooks, _ := os.Open("testing/WebhookExistsResp.json")
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