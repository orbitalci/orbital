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
	return &AppleDevProfile{}
}

type AppleDevProfile struct {
	// the zipped *.developerprofile secrets are retrieved from vault and set here
	keys []*ioshelper.AppleKeychain
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
	// todo: is this the best way?
	if wc.BuildTool == "xcode" {
		return true
	}
	return false
}

func (a *AppleDevProfile) GetEnv() []string {
	var envs []string
	// environment variables will be the contents of the apple keys to be imported to the keychain
	for _, key := range a.keys {
		for env, privateKeyData := range key.PrivateKeys {
			envs = append(envs, fmt.Sprintf("%s=%s", env, privateKeyData))
		}
		for env, mobileData := range key.MobileProvisions {
			envs = append(envs, fmt.Sprintf("%s=%s", env, mobileData))
		}
	}
	return envs
}

func (a *AppleDevProfile) MakeBashable(str string) []string {
	cmds := []string{"mkdir -p " + appleProfileDirec}
	pass := uuid.New().String()
	// delete old security profile if it exists
	cmds = append(cmds, "security delete-keychain ocelotty; echo \"deleting keychain whether it existed or not\"")
	//cmds = append(cmds, "if security list-keychains | grep ocelotty; then echo 'deleting' && security delete-keychain ocelotty; fi")
	// create a new security profile
	cmds = append(cmds, fmt.Sprintf("security create-keychain -p %s ocelotty && security unlock-keychain -p %s ocelotty", pass, pass))
	for _, key := range a.keys {
		for privKey := range key.PrivateKeys {
			// echo the private data to files
			cmds = append(cmds, fmt.Sprintf("echo ${%s} | base64 -D > %s/%s", privKey, appleProfileDirec, privKey))
			// add keys to ocelotty keychain
			cmds = append(cmds,  fmt.Sprintf("security import %s/%s -k ocelotty -p %s -T /usr/bin/codesign -T /usr/bin/productsign", appleProfileDirec, privKey, pass))
		}
		provisioningDir := "${HOME}/Library/MobileDevice/Provisioning\\ Profiles"
		for mobile := range key.MobileProvisions {
			cmds = append(cmds, fmt.Sprintf("echo \"installing %s\"", mobile))
			cmds = append(cmds, fmt.Sprintf("echo ${%s} | base64 -D > %s/%s", mobile, provisioningDir, mobile))
		}
	}
	cmds = append(cmds, "security list-keychains -d user -s login.keychain-db ocelotty-db", "echo \"wrote dev profile to keychains\"")
	combined := strings.Join(cmds, " && ")
	return []string{combined}
}


//  $ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d
//  $ security unlock-keychain d
//  $ security unlock-keychain d -p jessi
//  $ security unlock-keychain d
//  $ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d
//  $ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d -p jessi
//  $ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d -P jessi
//  $ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d -P jessi -T /usr/bin/codesign -T /usr/bin/productsign d
//  $ security show-keychain-info d
//  $ security
//  $ security list-keychains
//  $ security show-keychain-info
//  $ security show-keychain-info d
//  $ security show-keychain-info login
//  $ security help
//  $ security delete-keychain e
//  $ security delete-keychain d
//  $ security create-keychain
//  $ security list-keychains
//  $ security list-keychains -d jesseshank -s login.keychain-db -s ocelot-db
//  $ security list-keychains -d jesseshank -s login.keychain-db ocelot-db
//  $ security list-keychains -d user -s login.keychain-db ocelot-db
//  $ security list-keychains
//  $ history | grep security
//
//  jenkins keyring bullshit
//  @Override
//      public void perform(@Nonnull Run<?, ?> run, @Nonnull FilePath workspace, @Nonnull Launcher launcher, @Nonnull TaskListener listener) throws InterruptedException, IOException {
//          DeveloperProfile dp = getProfile(run.getParent());
//          if (dp==null)
//              throw new AbortException("No Apple developer profile is configured");
//
//          // Note: keychain are usualy suffixed with .keychain. If we change we should probably clean up the ones we created
//          String keyChain = "jenkins-"+run.getParent().getFullName().replace('/', '-');
//          String keychainPass = UUID.randomUUID().toString();
//
//          ArgumentListBuilder args;
//
//          {// if the key chain is already present, delete it and start fresh
//              ByteArrayOutputStream out = new ByteArrayOutputStream();
//              args = new ArgumentListBuilder("security","delete-keychain", keyChain);
//              launcher.launch().cmds(args).stdout(out).join();
//          }
//
//
//          args = new ArgumentListBuilder("security","create-keychain");
//          args.add("-p").addMasked(keychainPass);
//          args.add(keyChain);
//          invoke(launcher, listener, args, "Failed to create a keychain");
//
//          args = new ArgumentListBuilder("security","unlock-keychain");
//          args.add("-p").addMasked(keychainPass);
//          args.add(keyChain);
//          invoke(launcher, listener, args, "Failed to unlock keychain");
//
//          final FilePath secret = getSecretDir(workspace, keychainPass);
//          secret.unzipFrom(new ByteArrayInputStream(dp.getImage()));
//
//          // import identities
//          for (FilePath id : secret.list("**/*.p12")) {
//  args = new ArgumentListBuilder("security","import");
//  args.add(id).add("-k",keyChain);
//  args.add("-P").addMasked(dp.getPassword().getPlainText());
//  args.add("-T","/usr/bin/codesign");
//  args.add("-T","/usr/bin/productsign");
//  args.add(keyChain);
//  invoke(launcher, listener, args, "Failed to import identity "+id);
//  }
//
//  {
//  // display keychain info for potential troubleshooting
//  args = new ArgumentListBuilder("security","show-keychain-info");
//  args.add(keyChain);
//  ByteArrayOutputStream output = invoke(launcher, listener, args, "Failed to show keychain info");
//  listener.getLogger().write(output.toByteArray());
//  }
//
//  // copy provisioning profiles
//  VirtualChannel ch = launcher.getChannel();
//  FilePath home = ch.call(new GetHomeDirectory());    // TODO: switch to FilePath.getHomeDirectory(ch) when we can
//  FilePath profiles = home.child("Library/MobileDevice/Provisioning Profiles");
//  profiles.mkdirs();
//
//  for (FilePath mp : secret.list("**/*.mobileprovision")) {
//  listener.getLogger().println("Installing  "+mp.getName());
//  mp.copyTo(profiles.child(mp.getName()));
//  }
//  }
