package handler

import "github.com/shankj3/ocelot/admin/models"

//TODO: this is the planned interface for other version control clients, I guess?
//TODO: originally I made this to make testing easier but didn't do it quite right
type Handler interface {
	SetMeUp(adminConfig *models.AdminConfig) error
	Walk() error
	GetFile(filePath string, fullRepoName string, commitHash string)
	CreateWebhook(webhookURL string) error
}