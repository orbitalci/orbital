package bitbucket

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
	"golang.org/x/oauth2"
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

func TestBitbucket_PostPRComment(t *testing.T) {
	vcsConf := &pb.VCSCreds{
		ClientId:     "VEhMhdw6uprevzh8Du",
		ClientSecret: "JtxmvZy3QR2dJQwbnwmVktCJa4jaVsJS",
		TokenURL:     "https://bitbucket.org/site/oauth2/access_token",
		AcctName:     "level11consulting",
		SubType:      pb.SubCredType_BITBUCKET,
		Identifier:   pb.SubCredType_BITBUCKET.String() + "_level11consulting",
	}
	client, token, err := GetBitbucketClient(vcsConf)
	if err != nil {
		t.Error(err)
	}
	err = client.PostPRComment("level11consulting/go-til", "2", "1234", true, 1234)
	if err != nil {
		t.Error(err)
	}
	ctx := context.Background()
	otoken := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	authCli := oauth2.NewClient(ctx, otoken)
	handler := GetBitbucketFromHttpClient(authCli)
	err = handler.PostPRComment("level11consulting/go-til", "2", "1234", false, 1234)
	if err != nil {
		t.Error(err)
	}

}

type MockHttpClient struct {
	Unmarshaler *jsonpb.Unmarshaler
	net.HttpClient
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

func (mhc MockHttpClient) PostUrlForm(url string, form url.Values) (*http.Response, error) {
	return nil, nil
}
