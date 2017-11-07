//Package ocenet contains HTTP related utility tools
package ocenet

import (
	"bufio"
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/ocelog"
	"io/ioutil"
	"net/http"
)

//HttpClient is a client containing a pre-authenticated http client as returned by
//golang's oauth2 clientcredentials package as well as a protobuf json unmarshaler
type HttpClient struct {
	AuthClient  *http.Client
	Unmarshaler *jsonpb.Unmarshaler
}

//GetUrl will perform a GET on the specified URL and return the appropriate protobuf response
func (hu HttpClient) GetUrl(url string, unmarshalObj proto.Message) {
	resp, err := hu.AuthClient.Get(url)
	if err != nil {
		ocelog.LogErrField(err).Error("can't get url ", url)
	}
	reader := bufio.NewReader(resp.Body)

	if err := hu.Unmarshaler.Unmarshal(reader, unmarshalObj); err != nil {
		ocelog.LogErrField(err).Error("failed to parse response from ", url)
	}
	defer resp.Body.Close()
}

//PostUrl will perform a post on the specified URL. It takes in a json formatted body
//and returns an (optional) protobuf response
func (hu HttpClient) PostUrl(url string, body string, unmarshalObj proto.Message) {
	bodyBytes := []byte(body)
	resp, err := hu.AuthClient.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		ocelog.LogErrField(err).Error("can't post to url ", url)
	}

	if unmarshalObj != nil {
		reader := bufio.NewReader(resp.Body)

		if err := hu.Unmarshaler.Unmarshal(reader, unmarshalObj); err != nil {
			ocelog.LogErrField(err).Error("failed to parse response from ", url)
		}
	} else {
		respBody, _ := ioutil.ReadAll(resp.Body)
		ocelog.Log().Debug(string(respBody))
	}

	defer resp.Body.Close()
}
