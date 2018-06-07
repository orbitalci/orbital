package build_signaler

import (
	//"testing"

	"github.com/golang/protobuf/proto"
)

type testProducer struct {
	messages []proto.Message
	topics []string
}

func (tp *testProducer) WriteProto(message proto.Message, topicName string) error {
	tp.messages = append(tp.messages, message)
	tp.topics = append(tp.topics, topicName)
	return nil
}

