package main

import (
	"github.com/shankj3/ocelot/admin/models"
	"github.com/pkg/errors"
)

//validator for all admin related stuff
type AdminValidator struct {}

func GetValidator() *AdminValidator {
	return &AdminValidator {}
}

//validates config and returns json formatted error
func(adminValidator AdminValidator) ValidateConfig(adminCreds *models.Credentials) (error) {
	if len(adminCreds.AcctName) == 0 {
		return errors.New("acctName must contain at least one letter")
	}
	if len(adminCreds.ClientId) == 0 {
		return errors.New("clientId must contain at least one letter")
	}
	if len(adminCreds.ClientSecret) == 0 {
		return errors.New("clientSecret must contain at least one letter")
	}
	if len(adminCreds.TokenURL) == 0 {
		return errors.New("tokenURL must contain at least one letter")
	}
	switch adminCreds.Type {
	case "bitbucket":
		return nil
	default:
		return errors.New("creds must be one of the following type: bitbucket")
	}
	return nil
}