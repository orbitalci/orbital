package build

const (
	DEFAULT = "build"
)

// GetTopics will return the list of topics that the werker should subscribe to as defined by
//   Right now, the only types that change the topic list are SSH and EXEC
func GetTopics(tags []string) []string {
	if len(tags) == 0 {
		return []string{DEFAULT}
	}
	if tags[0] == "" {
		return []string{DEFAULT}
	}
	var topics []string
	for _, tag := range tags {
		topics = append(topics, getTopicFromTag(tag))
	}
	return topics
}

func getTopicFromTag(tag string) string {
	if tag == "" {
		return "build"
	}
	return "build_" + tag
}

// DetermineTopic will return the correct topic based on the value of machineTag
func DetermineTopic(machineTag string) (topic string) {
	switch {
	case machineTag == "":
		return DEFAULT
	default:
		return getTopicFromTag(machineTag)
	}
}