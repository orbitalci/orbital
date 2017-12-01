package main

import (
	"gopkg.in/go-playground/validator.v9"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/ocelog"
)

//validator for all admin related stuff
type AdminValidator struct {
	Validate	*validator.Validate
}

func GetValidator() *AdminValidator {
	adminValidator := &AdminValidator {
		Validate: validator.New(),
	}
	adminValidator.Validate.RegisterValidation("validtype", typeValidation)
	return adminValidator
}

//validates config and returns json formatted error
func(adminValidator AdminValidator) ValidateConfig(adminConfig *models.Credentials) (string, error) {
	err := adminValidator.Validate.Struct(adminConfig)
	if err != nil {
		var errorMsg string
		for _, nestedErr := range err.(validator.ValidationErrors) {
			errorMsg = nestedErr.Field() + " is " + nestedErr.Tag()
			if nestedErr.Tag() == "validtype" {
				errorMsg = "type must be one of the following: bitbucket"
			}

			ocelog.Log().Warn(errorMsg)
		}
		return errorMsg, err
	}
	return "", nil
}

func typeValidation(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "bitbucket":
		return true
	}
	return false
}