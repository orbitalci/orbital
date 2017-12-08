package admin

import (
	"testing"
	"github.com/shankj3/ocelot/admin/models"
	"bitbucket.org/level11consulting/go-til/test"
)

func TestAdminValidator_ValidateConfig(t *testing.T) {
	v := GetValidator()
	noClientSecret := &models.Credentials{
		AcctName: "blah",
		ClientId: "blah2",
		TokenURL: "slkdjf",
		Type: "bitbucket",
	}

	err := v.ValidateConfig(noClientSecret)
	if err.Error() != "clientSecret is required" {
		t.Error(test.GenericStrFormatErrors("client secret", "clientSecret is required", err.Error()))
	}

	invalidCred := &models.Credentials{
		AcctName: "blah",
		ClientId: "blah2",
		TokenURL: "slkdjf",
		ClientSecret: "jsdlkfsdfjskdf",
		Type: "marianne",
	}

	wrongType := v.ValidateConfig(invalidCred)
	if wrongType.Error() != "creds must be one of the following type: bitbucket" {
		t.Error(test.GenericStrFormatErrors("credential type", "creds must be one of the following type: bitbucket", wrongType.Error()))
	}
}