package admin

import (
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"testing"
)

func TestAdminValidator_ValidateConfig(t *testing.T) {
	v := GetValidator()
	noClientSecret := &models.VCSCreds{
		AcctName: "blah",
		ClientId: "blah2",
		TokenURL: "slkdjf",
		Type:     "bitbucket",
	}

	err := v.ValidateConfig(noClientSecret)
	if err.Error() != "clientSecret is required" {
		t.Error(test.GenericStrFormatErrors("client secret", "clientSecret is required", err.Error()))
	}

	invalidCred := &models.VCSCreds{
		AcctName:     "blah",
		ClientId:     "blah2",
		TokenURL:     "slkdjf",
		ClientSecret: "jsdlkfsdfjskdf",
		Type:         "marianne",
	}

	wrongType := v.ValidateConfig(invalidCred)
	if wrongType.Error() != "creds must be one of the following type: bitbucket" {
		t.Error(test.GenericStrFormatErrors("credential type", "creds must be one of the following type: bitbucket", wrongType.Error()))
	}
}
