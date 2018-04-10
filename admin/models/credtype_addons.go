package models

// Subtypes will return all the SubCredTypes that are associated with that CredType. Will return nil if it is unknown
func (x CredType) Subtypes() []SubCredType {
	switch x {
	case CredType_VCS:
		return []SubCredType{SubCredType_BITBUCKET, SubCredType_GITHUB}
	case CredType_REPO:
		return []SubCredType{SubCredType_NEXUS, SubCredType_MAVEN, SubCredType_DOCKER}
	case CredType_K8S:
		return []SubCredType{SubCredType_KUBECONF}
	}
	// this shouldn't happen, unless a new CredType is added and not updated here.
	return nil
}