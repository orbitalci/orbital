

```
$ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d
$ security unlock-keychain d
$ security unlock-keychain d -p jessi
$ security unlock-keychain d
$ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d
$ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d -p jessi
$ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d -P jessi
$ security import developer/identities/1D44087880734D2C42EF1B4F9684743EA22968D7.p12  -k d -P jessi -T /usr/bin/codesign -T /usr/bin/productsign d
$ security show-keychain-info d
$ security
$ security list-keychains
$ security show-keychain-info
$ security show-keychain-info d
$ security show-keychain-info login
$ security help
$ security delete-keychain e
$ security delete-keychain d
$ security create-keychain
$ security list-keychains
$ security list-keychains -d jesseshank -s login.keychain-db -s ocelot-db
$ security list-keychains -d jesseshank -s login.keychain-db ocelot-db
$ security list-keychains -d user -s login.keychain-db ocelot-db
$ security list-keychains
$ history | grep security
```

```java
jenkins keyring bullshit
@Override
    public void perform(@Nonnull Run<?, ?> run, @Nonnull FilePath workspace, @Nonnull Launcher launcher, @Nonnull TaskListener listener) throws InterruptedException, IOException {
        DeveloperProfile dp = getProfile(run.getParent());
        if (dp==null)
            throw new AbortException("No Apple developer profile is configured");

        // Note: keychain are usualy suffixed with .keychain. If we change we should probably clean up the ones we created
        String keyChain = "jenkins-"+run.getParent().getFullName().replace('/', '-');
        String keychainPass = UUID.randomUUID().toString();

        ArgumentListBuilder args;

        {// if the key chain is already present, delete it and start fresh
            ByteArrayOutputStream out = new ByteArrayOutputStream();
            args = new ArgumentListBuilder("security","delete-keychain", keyChain);
            launcher.launch().cmds(args).stdout(out).join();
        }


        args = new ArgumentListBuilder("security","create-keychain");
        args.add("-p").addMasked(keychainPass);
        args.add(keyChain);
        invoke(launcher, listener, args, "Failed to create a keychain");

        args = new ArgumentListBuilder("security","unlock-keychain");
        args.add("-p").addMasked(keychainPass);
        args.add(keyChain);
        invoke(launcher, listener, args, "Failed to unlock keychain");

        final FilePath secret = getSecretDir(workspace, keychainPass);
        secret.unzipFrom(new ByteArrayInputStream(dp.getImage()));

        // import identities
        for (FilePath id : secret.list("**/*.p12")) {
args = new ArgumentListBuilder("security","import");
args.add(id).add("-k",keyChain);
args.add("-P").addMasked(dp.getPassword().getPlainText());
args.add("-T","/usr/bin/codesign");
args.add("-T","/usr/bin/productsign");
args.add(keyChain);
invoke(launcher, listener, args, "Failed to import identity "+id);
}

{
// display keychain info for potential troubleshooting
args = new ArgumentListBuilder("security","show-keychain-info");
args.add(keyChain);
ByteArrayOutputStream output = invoke(launcher, listener, args, "Failed to show keychain info");
listener.getLogger().write(output.toByteArray());
}

// copy provisioning profiles
VirtualChannel ch = launcher.getChannel();
FilePath home = ch.call(new GetHomeDirectory());    // TODO: switch to FilePath.getHomeDirectory(ch) when we can
FilePath profiles = home.child("Library/MobileDevice/Provisioning Profiles");
profiles.mkdirs();

for (FilePath mp : secret.list("**/*.mobileprovision")) {
listener.getLogger().println("Installing  "+mp.getName());
mp.copyTo(profiles.child(mp.getName()));
}
}
```