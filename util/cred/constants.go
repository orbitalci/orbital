package cred

import (
	"os"
	"sync"
)
var once sync.Once
var prefix string

func getPrefix() string {
	once.Do(func(){
		prefix = os.Getenv("PATH_PREFIX")
		if prefix == "" {
			prefix = ""
		} else {
			prefix = prefix + "/"
		}
	})
	return prefix
}


var (
	OcyConfigBase = getPrefix() +  "config/ocelot"
	StorageType =  OcyConfigBase + "/storagetype"

	PostgresCredLoc = OcyConfigBase + "/postgres"
	PostgresDatabaseName = PostgresCredLoc + "/db"
	PostgresLocation = PostgresCredLoc + "/location"
	PostgresPort = PostgresCredLoc + "/port"
	PostgresUsername = PostgresCredLoc + "/username"
	PostgresPasswordLoc = "secret/" + PostgresCredLoc
	PostgresPasswordKey = "clientsecret"

	FilesystemConfigLoc =  OcyConfigBase + "/filesystem"
	FilesystemDir = FilesystemConfigLoc + "/savedirec"
	ConfigPath = "creds"
)
