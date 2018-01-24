package hookhandler
//has necessary functions for running hookhandler in dev mode

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"io/ioutil"
	"os"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/go-til/consul"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	"github.com/hashicorp/vault/api"
)

type MockHookHandlerContext struct {
	HookHandlerContext //embedding hookhandler context cause didn't want to stub out getters/setters
}

func (hhc *MockHookHandlerContext) GetBitbucketClient(cfg *models.VCSCreds) (handler.VCSHandler, string, error) {
	mockVcsHandler := &MockVCSHandler{}
	return mockVcsHandler , "", nil
}

func (hhc *MockHookHandlerContext) GetRemoteConfig() cred.CVRemoteConfig {
	return &MockRemoteConfig{}
}

////mock remote config/////

type MockRemoteConfig struct {}

func (mrc *MockRemoteConfig) GetConsul()	*consul.Consulet {
	return nil
}
func (mrc *MockRemoteConfig) SetConsul(consul *consul.Consulet) {}
func (mrc *MockRemoteConfig) GetVault() ocevault.Vaulty {
	return &MockVaulty{}
}
func (mrc *MockRemoteConfig) SetVault(vault ocevault.Vaulty) {}
func (mrc *MockRemoteConfig) GetCredAt(path string, hideSecret bool, ocyType cred.OcyCredType) (map[string]cred.RemoteConfigCred, error) {
	mockMap := make(map[string]cred.RemoteConfigCred)
	mockVcsCreds := &models.VCSCreds{}

	mockMap["bitbucket/mariannefeng"] = mockVcsCreds
	return mockMap, nil
}
func (mrc *MockRemoteConfig) GetPassword(path string) (string, error) {
	return "", nil
}
func (mrc *MockRemoteConfig) AddCreds(path string, anyCred cred.RemoteConfigCred) (err error) {
	return nil
}

////mock vault////

type MockVaulty struct {}

func (mv *MockVaulty) AddUserAuthData(user string, data map[string]interface{}) (*api.Secret, error) {
	return nil, nil
}
func (mv *MockVaulty) GetUserAuthData(user string) (map[string]interface{}, error) {
	mockMap := make(map[string]interface{})
	return mockMap, nil
}
func (mv *MockVaulty) CreateToken(request *api.TokenCreateRequest) (token string, err error) {
	return "", nil
}
func (mv *MockVaulty) CreateThrowawayToken() (token string, err error) {
	return "", nil
}
func (mv *MockVaulty) CreateVaultPolicy() error {
	return nil
}

////mock vcs handler////

type MockVCSHandler struct {}

func (mvh *MockVCSHandler) Walk() error {
	return nil
}

//****WARNING**** this assumes you're running this inside of the hookhandler folder
func (mvh *MockVCSHandler) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	pwd, _ := os.Getwd()
	return ioutil.ReadFile(pwd + "/test-fixtures/dev-ocelot.yml")
}

func (mvh *MockVCSHandler) CreateWebhook(webhookURL string) error {
	return nil
}

func (mvh *MockVCSHandler) GetCallbackURL() string {
	return ""
}

func (mvh *MockVCSHandler) SetCallbackURL(callbackURL string) {}

func (mvh *MockVCSHandler) SetBaseURL(baseURL string) {}

func (mvh *MockVCSHandler) GetBaseURL() string {
	return ""
}

func (mvh *MockVCSHandler) FindWebhooks(getWebhookURL string) bool {
	return true
}