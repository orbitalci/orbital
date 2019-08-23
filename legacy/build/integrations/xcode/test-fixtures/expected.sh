mkdir -p /tmp/.appleProfs
security delete-keychain ocelotty; echo "deleting keychain whether it existed or not"
security create-keychain -p TESTPASS ocelotty && security unlock-keychain -p TESTPASS ocelotty
echo ${id1_OCY_p12} | base64 -D > /tmp/.appleProfs/id1.p12
security import /tmp/.appleProfs/id1.p12 -k ocelotty -P DEVTESTPASS -T /usr/bin/codesign -T /usr/bin/productsign
echo "installing prof1.mobileprovision"
echo ${prof1_OCY_mobileprovision} | base64 -D > ${HOME}/Library/MobileDevice/Provisioning\ Profiles/prof1.mobileprovision
echo "installing prof2.mobileprovision"
echo ${prof2_OCY_mobileprovision} | base64 -D > ${HOME}/Library/MobileDevice/Provisioning\ Profiles/prof2.mobileprovision
security list-keychains -d user -s login.keychain-db ocelotty-db
echo "wrote dev profile to keychains"