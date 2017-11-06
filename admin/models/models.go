package models

//TODO: get callback url from consul or something like it
const WebhookCallbackURL string = "https://radiant-mesa-23210.herokuapp.com//test"
const BuildFileName string = "README.md"

type AdminConfig struct {
	ConfigId     string `json:"configId,omitempty"`
	ClientId     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	TokenURL     string `json:"tokenURL,omitempty"`
	AcctName     string `json:"acctName,omitempty"`
}
