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
	// nexus stuff
	Nexus = RepoPath + "/%s/nexus"
	NexusUrlPath = Nexus + "/repourl"
	Docker = RepoPath + "/%s/docker"
	K8sPath = ConfigPath + "/k8s"
	Kubernetes = K8sPath + "/%s/k8s"
)