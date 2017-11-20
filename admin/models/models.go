package models


//TODO: get callback url from consul or something like it
const WebhookCallbackURL = "https://radiant-mesa-23210.herokuapp.com/test"
const BuildFileName = "ocelot.yml"
const ConfigFileName = "config.yml"

type AdminConfig struct {
	ConfigId     string
	ClientId     string `yaml:"clientId" validate:"required"`
	ClientSecret string `yaml:"clientSecret" validate:"required"`
	TokenURL     string `yaml:"tokenURL" validate:"required"`
	AcctName     string `yaml:"acctName" validate:"required"`
	Type		 string `yaml:"type" validate:"bitbucket|github"`
}

type ConfigYaml struct {
	Credentials map[string]AdminConfig	`yaml:"credentials"`
}
