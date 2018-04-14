package testutil


func CompareCredWrappers(credWrapA *CredWrapper, credWrapB *CredWrapper) bool {
	for ind, cred := range credWrapA.Vcs {
		credB := credWrapB.Vcs[ind]
		if cred.SubType != credB.SubType {
			return false
		}
		if cred.AcctName != credB.AcctName {
			return false
		}
		if cred.TokenURL != credB.TokenURL {
			return false
		}
		if cred.ClientSecret != credB.ClientSecret {
			return false
		}
		if cred.ClientId != credB.ClientId {
			return false
		}
		if cred.SshFileLoc != credB.SshFileLoc {
			return false
		}
	}
	return true
}

func CompareRepoCredWrappers(repoWrapA *RepoCredWrapper, repoWrapB *RepoCredWrapper) bool {
	for ind, cred := range repoWrapA.Repo {
		credB := repoWrapB.Repo[ind]
		if cred.SubType != credB.SubType {
			return false
		}
		if cred.Username != credB.Username {
			return false
		}
		if cred.AcctName != credB.AcctName {
			return false
		}
		if cred.Password != credB.Password {
			return false
		}
		//for name, url := range cred.RepoUrl {
		//todo: fix this
		//}
	}
	return true
}

func CompareAllCredWrappers(allWrapA *AllCredsWrapper, allWrapB *AllCredsWrapper) bool {
	if repoMatches := CompareRepoCredWrappers(allWrapA.RepoCreds, allWrapB.RepoCreds); !repoMatches {
		return false
	}
	if vcsMatches := CompareCredWrappers(allWrapA.VcsCreds, allWrapB.VcsCreds); !vcsMatches {
		return false
	}
	return true
}
