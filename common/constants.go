package common

const BuildFileName = "ocelot.yml"

var BitbucketEvents = []string{"repo:push", "pullrequest:approved", "pullrequest:updated"}

var (
	OcyConfigBase = GetPrefix() + "config/ocelot"
	StorageType   = OcyConfigBase + "/storagetype"

	PostgresCredLoc      = OcyConfigBase + "/postgres"
	PostgresDatabaseName = PostgresCredLoc + "/db"
	PostgresLocation     = PostgresCredLoc + "/location"
	PostgresPort         = PostgresCredLoc + "/port"
	PostgresUsername     = PostgresCredLoc + "/username"
	PostgresPasswordLoc  = "secret/data/" + PostgresCredLoc
	PostgresPasswordKey  = "clientsecret"

	FilesystemConfigLoc = OcyConfigBase + "/filesystem"
	FilesystemDir       = FilesystemConfigLoc + "/savedirec"
	ConfigPath          = GetPrefix() + "creds"
)
