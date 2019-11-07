package ioshelper

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/level11consulting/orbitalci/build/helpers/serde"
)

func GetAndEncodeDevFolder(t *testing.T) map[string]string {
	profiles := make(map[string]string)
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	aProfile, err := ioutil.ReadFile(filepath.Join(dir, "test-fixtures/developer/profiles/prof1.mobileprovision"))
	if err != nil {
		t.Fatal(err)
	}
	encoded := serde.BitzToBase64(aProfile)
	profiles["prof1.mobileprovision"] = encoded
	prof2, err := ioutil.ReadFile(filepath.Join(dir, "test-fixtures/developer/profiles/prof2.mobileprovision"))
	if err != nil {
		t.Fatal(err)
	}
	profiles["prof2.mobileprovision"] = serde.BitzToBase64(prof2)
	identity, err := ioutil.ReadFile(filepath.Join(dir, "/test-fixtures/developer/identities/id1.p12"))
	if err != nil {
		t.Fatal(err)
	}
	profiles["id1.p12"] = serde.BitzToBase64(identity)

	return profiles
}

func GetZipAndPw(t *testing.T) (zipdata []byte, pw string) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	zipdata, err := ioutil.ReadFile(filepath.Join(dir, "test-fixtures/test.zip"))
	if err != nil {
		t.Fatal(err)
	}
	return zipdata, "pw"
}
