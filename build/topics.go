package build

import (
	"strings"

	"github.com/shankj3/ocelot/models"
)

const (
	DEFAULT = "build"
	SSH     = "build_ssh"
)

// GetTopics will return the list of topics that the werker should subscribe to as defined by his werk type
//  Right now, the only type that changes the topic list is SSH.
func GetTopics(werkType models.WerkType) []string {
	switch werkType {
	case models.SSH:
		return []string{SSH}
	default:
		return []string{DEFAULT}
	}
}

// DetermineTopic will return the correct topic based on the value of buildTool
func DetermineTopic(buildTool string) (topic string) {
	switch {
	case strings.Contains(buildTool, "xcode"):
		return SSH
	default:
		return DEFAULT
	}
}