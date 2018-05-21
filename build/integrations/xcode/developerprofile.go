package xcode

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shankj3/ocelot/common/helpers/ioshelper"
	"github.com/shankj3/ocelot/models/pb"
)

const (
	appleProfileDirec   = "/tmp/.appleProfs"
)


func Create() *AppleDevProfile {
	return &AppleDevProfile{joiner: " && ", pass: uuid.New().String()}
}

type AppleDevProfile struct {
	// the zipped *.developerprofile secrets are retrieved from vault and set here
	keys []*ioshelper.AppleKeychain
	joiner string
	pass   string
}

func (a *AppleDevProfile) String() string {
	return "apple dev profile integration"
}

func (a *AppleDevProfile) SubType() pb.SubCredType {
	return pb.SubCredType_DEVPROFILE
}

func (a *AppleDevProfile) GenerateIntegrationString(creds []pb.OcyCredder) (contents string, err error) {
	for _, cred := range creds {
		keyc := ioshelper.NewKeychain()
		iosCred := cred.(*pb.AppleCreds)
		err = json.Unmarshal(iosCred.AppleSecrets, keyc)
		if err != nil {
			return
		}
		a.keys = append(a.keys, keyc)
	}
	return
}

func (a *AppleDevProfile) IsRelevant(wc *pb.BuildConfig) bool {
	// todo: make apple dev profile more robust / not f*** with any existing keychains and then use this. for now people will have to be logged in to the mac build node
	// todo: is this the best way?
	//if wc.BuildTool == "xcode" {
	//	return true
	//}
	return false
}

func (a *AppleDevProfile) GetEnv() []string {
	var envs []string
	// environment variables will be the contents of the apple keys to be imported to the keychain
	for _, key := range a.keys {
		for env, privateKeyData := range key.PrivateKeys {
			envs = append(envs, fmt.Sprintf("%s=%s", makeEnvValid(env), privateKeyData))
		}
		for env, mobileData := range key.MobileProvisions {
			envs = append(envs, fmt.Sprintf("%s=%s", makeEnvValid(env), mobileData))
		}
	}
	return envs
}

// makeEnvValid will replace all the dots in the env var name
func makeEnvValid(envName string) string{
	return strings.Replace(envName, ".", "_OCY_", -1)
}

func (a *AppleDevProfile) MakeBashable(str string) []string {
	cmds := []string{"mkdir -p " + appleProfileDirec}
	// delete old security profile if it exists
	cmds = append(cmds, "security delete-keychain ocelotty; echo \"deleting keychain whether it existed or not\"")
	//cmds = append(cmds, "if security list-keychains | grep ocelotty; then echo 'deleting' && security delete-keychain ocelotty; fi")
	// create a new security profile
	cmds = append(cmds, fmt.Sprintf("security create-keychain -p %s ocelotty && security unlock-keychain -p %s ocelotty", a.pass, a.pass))
	for _, key := range a.keys {
		for privKey := range key.PrivateKeys {
			// echo the private data to files
			cmds = append(cmds, fmt.Sprintf("echo ${%s} | base64 -D > %s/%s", makeEnvValid(privKey), appleProfileDirec, privKey))
			// add keys to ocelotty keychain
			cmds = append(cmds,  fmt.Sprintf("security import %s/%s -k ocelotty -P %s -T /usr/bin/codesign -T /usr/bin/productsign", appleProfileDirec, privKey, key.DevProfilePassword))
		}
		provisioningDir := "${HOME}/Library/MobileDevice/Provisioning\\ Profiles"
		for mobile := range key.MobileProvisions {
			cmds = append(cmds, fmt.Sprintf("echo \"installing %s\"", mobile))
			cmds = append(cmds, fmt.Sprintf("echo ${%s} | base64 -D > %s/%s", makeEnvValid(mobile), provisioningDir, mobile))
		}
	}
	cmds = append(cmds, "security list-keychains -d user -s login.keychain-db ocelotty-db", "echo \"wrote dev profile to keychains\"")
	combined := strings.Join(cmds, a.joiner)
	return []string{combined}
}
