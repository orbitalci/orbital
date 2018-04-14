package credentials


// GetVcsCreds will retrieve a VCSCred for account name / bitbucket vcs type
func GetVcsCreds(store storage.CredTable, repoFullName string, remoteConfig cred.CVRemoteConfig) (*models.VCSCreds, error) {
	acctName, _, err := GetAcctRepo(repoFullName)
	if err != nil {
		return nil, err
	}
	identifier, err := models.CreateVCSIdentifier(models.SubCredType_BITBUCKET, acctName)
	if err != nil {
		return nil, err
	}
	bbCreds, err := remoteConfig.GetCred(store, models.SubCredType_BITBUCKET, identifier, acctName, false)
	if err != nil {
		return nil, err
	}
	vcs, ok := bbCreds.(*models.VCSCreds)
	if !ok {
		return nil, errors.New("could not cast as vcs creds")
	}
	return vcs, err
}
