package storage

import (
	"bytes"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
	"testing"
	"bitbucket.org/level11consulting/go-til/test"
)


func testGetHomeDirec() string{
	direc, err := homedir.Dir()
	if err != nil {
		panic("need home direc capability")
	}
	return direc
}
var hash = "12345678"
var HomeDirec = testGetHomeDirec()

var filebuildstorages = []struct{
	initSaveDirec   string
	actualSaveLoc string
}{
	{"", filepath.Join(HomeDirec, ".ocelot", "build-output", hash)},
	{"/tmp/test/one", filepath.Join("/tmp", "test", "one", hash)},
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) {return false, nil}
	return true, err
}


func TestFileBuildStorage(t *testing.T) {
	testBytes := []byte("woooooooeooooeooeeeoeo!!!!!")

	for _, fb := range filebuildstorages {
		fbs := &FileBuildStorage{
			saveDirec: fb.initSaveDirec,
		}
		fbs.setup()

		err := fbs.Store(hash, testBytes)
		if err != nil {
			t.Fatal(err)
		}
		if filep := fbs.getTempFile(hash); filep != fb.actualSaveLoc {
			t.Error(test.StrFormatErrors("file path", fb.actualSaveLoc, filep))
		}

		actualData, err := fbs.Retrieve(hash)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actualData, testBytes){
			t.Error(test.GenericStrFormatErrors("file data", string(testBytes), string(actualData)))
		}
		fbs.Clean()
		exists, _ := exists(fb.actualSaveLoc)
		if exists {
			t.Error("save directory should not exist. path: ", fb.actualSaveLoc)
		}
	}


}