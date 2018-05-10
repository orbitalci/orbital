package build

import (
	"strings"

	"github.com/shankj3/ocelot/models"
)

const (
	DEFAULT = "build"
	BARE    = "build_baremetal"
)

// GetTopics will return the list of topics that the werker should subscribe to as defined by his werk type
//   Right now, the only types that change the topic list are SSH and EXEC
func GetTopics(werkType models.WerkType) []string {
	switch werkType {
	case models.SSH, models.Exec:
		return []string{BARE}
	default:
		return []string{DEFAULT}
	}
}

// DetermineTopic will return the correct topic based on the value of buildTool
func DetermineTopic(buildTool string) (topic string) {
	switch {
	case strings.Contains(buildTool, "xcode"):
		return BARE
	default:
		return DEFAULT
	}
}