package nsqpb

import (
	"encoding/json"
	"fmt"
	"github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util/ocelog"
	"net/http"
)

const (
	PRTopic   = "pr"
	PushTopic = "push"

	Build = "build"
)

var SupportedTopics = [2]string{PushTopic, PRTopic}

// extends proto.Message interface
type BundleProtoMessage interface {
	GetCheckoutHash() string
	Reset()
	String() string
	ProtoMessage()
}



func TopicsUnmarshalObj(topic string) BundleProtoMessage {
	switch topic {
	case PRTopic:   return &protos.PRBuildBundle{}
	case PushTopic: return &protos.PushBuildBundle{}
	default:        return nil
	}
}

type Topics struct {
	topics []string
}

func GetTopics(nsqdlookupHostPort string) (*Topics, error) {
	nsqdlookupAddr := fmt.Sprintf("http://%s/topics", nsqdlookupHostPort)
	resp, err := http.Get(nsqdlookupAddr)
	if err != nil {
		return nil, err
	}
	tops := &Topics{}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(tops); err != nil {
		return nil, err
	}
	return tops, nil
}

func LookupTopic(nsqdLookupHostPort string, topic string) bool {
	nsqdLookupAddr := fmt.Sprintf("http://%s/lookup?topic=%s", nsqdLookupHostPort, topic)
	resp, err := http.Get(nsqdLookupAddr)
	if err != nil {
		ocelog.IncludeErrField(err).Fatalf("Error on looking up topic! %s", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return false
	}
	// todo: check for other headers?
	return true
}