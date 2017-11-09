package deserialize

import (
	"testing"
	"io/ioutil"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util"
	pb "github.com/shankj3/ocelot/protos/out"
	"bytes"
)

const TestConfig = "test/testconfig.yml"
const TestOcelot = "test/ocelot.yml"
const TestRepos = "test/repo.json"

func TestDeserializer_YAMLToStruct(t *testing.T) {
	testFile, _ := ioutil.ReadFile(TestConfig)
	d := New()

	configYaml := &models.ConfigYaml{}
	d.YAMLToStruct(testFile, configYaml)

	if len(configYaml.Credentials) > 1 {
		t.Errorf("expected credential %d, got %d", 1, len(configYaml.Credentials))
	}

	for key, value := range configYaml.Credentials {
		if key != "marianne" {
			t.Error(util.StrFormatErrors("credential key", "marianne", key))
		}
		if value.ClientId != "abc" {
			t.Error(util.StrFormatErrors("client id", "abc", value.ClientId))
		}
		if value.ClientSecret != "abcd" {
			t.Error(util.StrFormatErrors("client secret", "abcd", value.ClientSecret))
		}
		if value.TokenURL != "abcdef" {
			t.Error(util.StrFormatErrors("token url", "abcdef", value.TokenURL))
		}
		if value.AcctName != "abcdefg" {
			t.Error(util.StrFormatErrors("account name", "abcdefg", value.AcctName))
		}
	}
}

func TestDeserializer_YAMLToProto(t *testing.T) {
	testOcelot, _ := ioutil.ReadFile(TestOcelot)
	d := New()
	ocelot := &pb.BuildConfig{}
	d.YAMLToProto(testOcelot, ocelot)

	if ocelot.Image != "test" {
		t.Error(util.StrFormatErrors("ocelot image", "test", ocelot.Image))
	}
	if len(ocelot.DockerPackages) != 2 {
		t.Error(util.IntFormatErrors("docker package list size", 2, len(ocelot.DockerPackages)))
	}
	if ocelot.BeforeStages.Env != nil {
		t.Error(util.StrFormatErrors("before stages environment", "", ocelot.BeforeStages.Env[0]))
	}
	if ocelot.BeforeStages.Script[0] != "sh echo \"hello\"" {
		t.Error(util.StrFormatErrors("before stages first script", "sh echo \"hello\"", ocelot.BeforeStages.Script[0]))
	}
	if ocelot.AfterStages.Env[0] != "BUILD_DEBUG=0" {
		t.Error(util.StrFormatErrors("after stages first environment", "BUILD_DEBUG=0", ocelot.AfterStages.Env[0]))
	}
	//can we assume parsing looks good if the above values have been set or do I have to write it for all the fields
}

func TestDeserializer_JSONToProto(t *testing.T) {
	repositories := &pb.PaginatedRepository{}
	testRepo, _ := ioutil.ReadFile(TestRepos)
	d := New()
	d.JSONToProto(ioutil.NopCloser(bytes.NewReader(testRepo)), repositories)

	if repositories.Pagelen != 10 {
		t.Error(util.IntFormatErrors("repository page len", 10, int(repositories.Pagelen)))
	}
	if repositories.Page != 1 {
		t.Error(util.IntFormatErrors("repository current page", 1, int(repositories.Page)))
	}
	if repositories.Size != 1 {
		t.Error(util.IntFormatErrors("repository response size", 1, int(repositories.Size)))
	}
	if len(repositories.Values) != 1 {
		t.Error(util.IntFormatErrors("repository values length", 1, len(repositories.Values)))
	}
	if repositories.Values[0].Name != "test-ocelot" {
		t.Error(util.StrFormatErrors("repository name", "test-ocelot", repositories.Values[0].Name))
	}
	if repositories.Values[0].FullName != "mariannefeng/test-ocelot" {
		t.Error(util.StrFormatErrors("repository full name", "mariannefeng/test-ocelot", repositories.Values[0].FullName))
	}
	if repositories.Values[0].Type != "repository" {
		t.Error(util.StrFormatErrors("repository type", "repository", repositories.Values[0].Type))
	}
	if repositories.Values[0].Links.Hooks.Href != "https://api.bitbucket.org/2.0/repositories/mariannefeng/test-ocelot/hooks" {
		t.Error(util.StrFormatErrors("webhook", "https://api.bitbucket.org/2.0/repositories/mariannefeng/test-ocelot/hooks", repositories.Values[0].Links.Hooks.Href))
	}
}