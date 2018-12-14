package messageservice

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
)

var tests = []struct {
	name        string
	tagStrInput string
	expected    []string
}{
	{"empty", "", []string{DEFAULT}},
	{"notempty", "ios,baremetal", []string{"build_ios", "build_baremetal"}},
	{"longs", "ios,baremetal,1728,", []string{"build_ios", "build_baremetal", "build_1728", DEFAULT}},
}

func TestGetTopics(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitTags := strings.Split(tt.tagStrInput, ",")
			topics := GetTopics(splitTags)
			if diff := deep.Equal(tt.expected, topics); diff != nil {
				t.Error(diff)
			}
		})
	}
}
