package cred

const (
	OcyConfigBase = "config/ocelot"
	StorageType = OcyConfigBase + "/storagetype"

	PostgresCredLoc =  OcyConfigBase + "/postgres"
	PostgresDatabaseName = PostgresCredLoc + "/db"
	PostgresLocation = PostgresCredLoc + "/location"
	PostgresPort = PostgresCredLoc + "/port"
	PostgresUsername = PostgresCredLoc + "/username"
	PostgresPasswordLoc = "secret/" + PostgresCredLoc
	PostgresPasswordKey = "clientsecret"

	FilesystemConfigLoc = OcyConfigBase + "/filesystem"
	FilesystemDir = FilesystemConfigLoc + "/savedirec"


	ConfigPath = "creds"
	VCSPath = ConfigPath + "/vcs"
	RepoPath = ConfigPath + "/repo"
)