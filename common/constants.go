package common

const BuildFileName = "ocelot.yml"

var BitbucketEvents = []string{"repo:push", "pullrequest:approved", "pullrequest:updated"}

var (
	OcyConfigBase = GetPrefix() + "config/ocelot"
	StorageType   = OcyConfigBase + "/storagetype"

	//// For configuring how we get Postgres credentials
	//VaultConf           = OcyConfigBase + "/vault"
	//VaultDBSecretEngine = VaultConf + "/secretbackend"
	//VaultRoleName       = VaultConf + "/rolename"

	PostgresCredLoc      = OcyConfigBase + "/postgres"
	PostgresDatabaseName = PostgresCredLoc + "/db"
	PostgresLocation     = PostgresCredLoc + "/location"
	PostgresPort         = PostgresCredLoc + "/port"
	PostgresUsername     = PostgresCredLoc + "/username"

	// For static DB connection info in vault
	PostgresPasswordLoc = "secret/data/" + PostgresCredLoc
	PostgresPasswordKey = "clientsecret"

	// For dynamic DB connection info in vault
	PostgresVaultConf          = PostgresCredLoc + "/vault"
	PostgresVaultRoleName      = PostgresVaultConf + "/rolename"
	PostgresVaultSecretsEngine = PostgresVaultConf + "/secretsengine"

	FilesystemConfigLoc = OcyConfigBase + "/filesystem"
	FilesystemDir       = FilesystemConfigLoc + "/savedirec"
	ConfigPath          = GetPrefix() + "creds"
)
