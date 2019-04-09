package ioshelper

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/level11consulting/ocelot/build/helpers/serde"
)

func NewKeychain() *AppleKeychain {
	return &AppleKeychain{PrivateKeys: make(map[string]string), MobileProvisions: make(map[string]string)}
}

// appleKeychain holds the identities for ever added apple developer profile
type AppleKeychain struct {
	// PrivateKeys are the *.p12 extensions
	PrivateKeys map[string]string `json:"privateKeys,omitempty"`
	// MobileProvisions are the *.mobileprovision files
	MobileProvisions map[string]string `json:"mobileProvisions,omitempty"`
	// DevProfilePassword is the password you entered when you exported the account
	DevProfilePassword string `json:"devProfilePassword,omitempty"`
	// right now accountsKeychain and accountsPlist aren't used, so keeping private
	// accounts.keychain
	accountsKeychain string
	//accounts.plist
	accountsPlist string
}

// UnpackAppleDevAccount will take in the byte array of a sent zip file, assign each value to the appropriate
// fields in AppleKeychain, and return the marshalaled json
func UnpackAppleDevAccount(zipBytes []byte, devProfilePw string) ([]byte, error) {
	//unpack all that zippy goodness
	zipData := bytes.NewReader(zipBytes)
	keyc := NewKeychain()
	err := keyc.GetSecretsFromZip(zipData, devProfilePw)
	if err != nil {
		return nil, err
	}
	bittyChain, err := json.Marshal(keyc)
	if err != nil {
		return nil, err
	}
	return bittyChain, err
}

type SizeReader interface {
	io.ReaderAt
	Size() int64
}

func (a *AppleKeychain) GetSecretsFromZip(profileReadre SizeReader, profPass string) error {
	// todo: ... what ewre you thinkin with this DevProfilePassword?
	a.DevProfilePassword = profPass
	profileZip, err := zip.NewReader(profileReadre, profileReadre.Size())
	if err != nil {
		return err
	}
	for _, secretFile := range profileZip.File {
		if secretFile.FileInfo().IsDir() {
			continue
		}
		contents, err := secretFile.Open()
		if err != nil {
			return err
		}
		bytec, err := ioutil.ReadAll(contents)
		if err != nil {
			return err
		}
		fn := secretFile.FileInfo().Name()
		switch {
		case strings.Contains(fn, ".mobileprovision"):
			a.MobileProvisions[fn] = serde.BitzToBase64(bytec)
		case strings.Contains(fn, ".p12"):
			a.PrivateKeys[fn] = serde.BitzToBase64(bytec)
		case strings.Contains(fn, ".keychain"):
			a.accountsKeychain = serde.BitzToBase64(bytec)
		case strings.Contains(fn, ".plist"):
			a.accountsPlist = serde.BitzToBase64(bytec)
		default:
			return errors.New("unsupported file: " + fn)
		}
	}
	return nil
}
