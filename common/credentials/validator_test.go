package credentials

import (
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/models/pb"
	"testing"
)

func TestAdminValidator_ValidateConfig(t *testing.T) {
	v := GetValidator()
	noClientSecret := &pb.VCSCreds{
		AcctName: "blah",
		ClientId: "blah2",
		TokenURL: "slkdjf",
		SubType:  pb.SubCredType_BITBUCKET,
	}

	err := v.ValidateConfig(noClientSecret)
	if err.Error() != "clientSecret is required" {
		t.Error(test.GenericStrFormatErrors("client secret", "clientSecret is required", err.Error()))
	}

	invalidCred := &pb.VCSCreds{
		AcctName:     "blah",
		ClientId:     "blah2",
		TokenURL:     "slkdjf",
		ClientSecret: "jsdlkfsdfjskdf",
		SubType:      pb.SubCredType_NIL_SCT,
	}

	wrongType := v.ValidateConfig(invalidCred)
	if wrongType.Error() != "SUB CRED TYPE WAS NOT INSTANTIATED PROPERLY" {
		t.Error(test.GenericStrFormatErrors("credential type", "creds must be one of the following type: bitbucket", wrongType.Error()))
	}

	invalidCred = &pb.VCSCreds{
		AcctName:     "blah",
		ClientId:     "blah2",
		TokenURL:     "slkdjf",
		ClientSecret: "jsdlkfsdfjskdf",
		SubType:      pb.SubCredType_GITHUB,
	}

	wrongType = v.ValidateConfig(invalidCred)
	if wrongType.Error() != "creds must be one of the following type: bitbucket" {
		t.Error(test.GenericStrFormatErrors("credential type", "creds must be one of the following type: bitbucket", wrongType.Error()))
	}
}
