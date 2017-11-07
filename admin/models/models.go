package models


//TODO: get callback url from consul or something like it
const WebhookCallbackURL string = "https://radiant-mesa-23210.herokuapp.com//test"
const BuildFileName string = "README.md"

type AdminConfig struct {
	ConfigId     string
	ClientId     string `validate:"required"`
	ClientSecret string `validate:"required"`
	TokenURL     string `validate:"required"`
	AcctName     string `validate:"required"`
}
