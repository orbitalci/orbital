package common

const BuildFileName = "ocelot.yml"

var BitbucketEvents = []string{"repo:push", "pullrequest:approved", "pullrequest:updated"}

var (
	OcyConfigBase = GetPrefix() + "config/ocelot"
	StorageType   = OcyConfigBase + "/storagetype"

	// For configuring how we get Postgres credentials
	VaultConf           = OcyConfigBase + "/vault"
	VaultDBSecretEngine = VaultConf + "/secretbackend"
	VaultRoleName       = VaultConf + "/rolename"

	// For static DB connection info in vault
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
