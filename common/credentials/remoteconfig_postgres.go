package credentials

import (
	"fmt"
	"strconv"

	"github.com/level11consulting/ocelot/common"
	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"
)

// getForPostgres Reads configuration from Consul
func (rc *RemoteConfig) getForPostgres() (*StorageCreds, error) {

	// Credentials for Postgres may come from 2 mutually exclusive areas in Vault
	// Either creds are static, and stored as a KV secret in vault
	// Or creds are dynamic, and managed by vault's database secret engine

	// The desired default behavior is to use static secrets. Don't error out if Vault settings aren't in Consul
	postgresVaultConf, _ := rc.Consul.GetKeyValues(common.PostgresVaultConf)

	postgresVaultBackend := "kv" // The default backend is "kv" if left unconfigured in Consul
	postgresVaultRole := ""
	for _, vconf := range postgresVaultConf {
		switch vconf.Key {
		case common.PostgresVaultSecretsEngine:
			if dbConfigType := string(vconf.Value); dbConfigType != "kv" && dbConfigType != "database" {
				return &StorageCreds{}, errors.New("Unsupported Vault DB Secret Engine: " + dbConfigType)
			}
			postgresVaultBackend = string(vconf.Value)
		case common.PostgresVaultRoleName:
			postgresVaultRole = string(vconf.Value)
		}
	}

	storeConfig := &StorageCreds{}

	kvconfig, err := rc.Consul.GetKeyValues(common.PostgresCredLoc)
	if len(kvconfig) == 0 || err != nil {
		errorMsg := fmt.Sprintf("unable to get postgres creds from consul")
		if err != nil {
			return storeConfig, errors.Wrap(err, errorMsg)
		}
		return nil, errors.New(errorMsg)
	}

	for _, pair := range kvconfig {
		switch pair.Key {
		case common.PostgresDatabaseName:
			storeConfig.DbName = string(pair.Value)
		case common.PostgresLocation:
			storeConfig.Location = string(pair.Value)
		case common.PostgresPort:
			// todo: check for err
			storeConfig.Port, _ = strconv.Atoi(string(pair.Value))
		}
	}

	switch postgresVaultBackend {
	case "kv":
		ocelog.Log().Info("Static Postgres creds")
		kvconfig, err := rc.Consul.GetKeyValues(common.PostgresCredLoc)
		if len(kvconfig) == 0 || err != nil {
			errorMsg := fmt.Sprintf("unable to get postgres location from consul")
			if err != nil {
				return &StorageCreds{}, errors.Wrap(err, errorMsg)
			}
			return &StorageCreds{}, errors.New(errorMsg)
		}
		for _, pair := range kvconfig {
			switch pair.Key {
			case common.PostgresUsername:
				storeConfig.User = string(pair.Value)
			}
		}

		secrets, err := rc.Vault.GetVaultData(common.PostgresPasswordLoc)
		if len(secrets) == 0 || err != nil {
			errorMsg := fmt.Sprintf("unable to get postgres password from consul")
			if err != nil {
				return &StorageCreds{}, errors.Wrap(err, errorMsg)
			}
			return &StorageCreds{}, errors.New(errorMsg)
		}

		// making name clientsecret because i feel like there must be a way for us to genericize remoteConfig
		storeConfig.Password = fmt.Sprintf("%v", secrets[common.PostgresPasswordKey])

	case "database":
		ocelog.Log().Info("Dynamic Postgres creds")
		secrets, err := rc.Vault.GetVaultSecret(fmt.Sprintf("database/creds/%s", postgresVaultRole))
		if err != nil {
			errorMsg := fmt.Sprintf("unable to get dynamic postgres creds from vault")
			if err != nil {
				return &StorageCreds{}, errors.Wrap(err, errorMsg)
			}
			return &StorageCreds{}, errors.New(errorMsg)
		}

		storeConfig.User = fmt.Sprintf("%v", secrets.Data["username"].(string))
		storeConfig.Password = fmt.Sprintf("%v", secrets.Data["password"].(string))

		//ocelog.Log().Debugf("Dynamic postgres creds from Vault: %v", secrets)

		// Kick of a lease renewal goroutine
		go rc.Vault.RenewLeaseForever(secrets)
	}

	return storeConfig, nil
}
