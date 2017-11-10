//Package ocenet contains HTTP related utility tools
package ocenet

import (
	"bufio"
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/util/ocelog"
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
func (hu HttpClient) GetUrl(url string, unmarshalObj proto.Message) error {
	resp, err := hu.AuthClient.Get(url)
	defer resp.Body.Close()
	if err != nil {
		ocelog.IncludeErrField(err).Error("can't get url ", url)
		return err
	}
	reader := bufio.NewReader(resp.Body)

	if err := hu.Unmarshaler.Unmarshal(reader, unmarshalObj); err != nil {
		ocelog.IncludeErrField(err).Error("failed to parse response from ", url)
		return err
	}

	return nil
}

//PostUrl will perform a post on the specified URL. It takes in a json formatted body
//and returns an (optional) protobuf response
func (hu HttpClient) PostUrl(url string, body string, unmarshalObj proto.Message) error {
	bodyBytes := []byte(body)
	resp, err := hu.AuthClient.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	defer resp.Body.Close()

	if err != nil {
		ocelog.IncludeErrField(err).Error("can't post to url ", url)
		return err
	}

	if unmarshalObj != nil {
		reader := bufio.NewReader(resp.Body)

		if err := hu.Unmarshaler.Unmarshal(reader, unmarshalObj); err != nil {
			ocelog.IncludeErrField(err).Error("failed to parse response from ", url)
			return err
		}
	} else {
		respBody, _ := ioutil.ReadAll(resp.Body)
		ocelog.Log().Debug(string(respBody))
	}

	return nil
}
