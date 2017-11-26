package storage

import (
	"github.com/mitchellh/go-homedir"
	"github.com/shankj3/ocelot/util/ocelog"
	"io/ioutil"
	"os"
	"path/filepath"
	"io"
)

// Interface for any storage type that we pick (mongo, mysql, filesystem..)
// BuildOutputStorage is for storing build output from docker container.
type BuildOutputStorage interface {
	Retrieve(gitHash string) ([]byte, error)
	Store(gitHash string, data []byte) error
	RetrieveReader(gitHash string) (io.Reader, error)
}

// FileBuildStorage is an implementation of BuildOutputStorage that is for filesystem.
type FileBuildStorage struct {
	saveDirec string
}

// NewFileBuildStorage will return an initialized FileBuildStorage with the saveDirec added to the
// filesystem if it isn't already. will fatally exit if cannot create directory.
// if saveDir == "", will create path at `~/.ocelot/build-output`
func NewFileBuildStorage(saveDir string) (f *FileBuildStorage) {
	f = &FileBuildStorage{saveDirec: saveDir}
	f.setup()
	return
}

func (f *FileBuildStorage) setup(){
	var err error
	if f.saveDirec == "" {
		direc, err := homedir.Expand(filepath.Join("~", ".ocelot", "build-output"))
		if err != nil {
			ocelog.IncludeErrField(err).Fatal("could not init build storage")
		}
		f.saveDirec = direc
	}
	if err = os.MkdirAll(f.saveDirec, 0777); err != nil {
		ocelog.IncludeErrField(err).Fatal("could not init build storage")
	}
}

// Store build data to a file at <saveDirec>/hash
func (f *FileBuildStorage) Store(gitHash string, data []byte) (err error) {
	fp := f.GetTempFile(gitHash)
	if err = ioutil.WriteFile(fp, data, 0600); err != nil {
		return
	}
	return
}

// retrieve build data from filesystem
func (f *FileBuildStorage) Retrieve(gitHash string) (data []byte, err error) {
	fp := f.GetTempFile(gitHash)
	data, err = ioutil.ReadFile(fp)
	return
}

// retrieve build data from filesystem as io.Reader
func (f *FileBuildStorage) RetrieveReader(gitHash string) (read io.Reader, err error){
	fp := f.GetTempFile(gitHash)
	read, err = os.Open(fp)
	return
}


func (f FileBuildStorage) GetTempFile(gitHash string) string {
	fp := filepath.Join(f.saveDirec, gitHash)
	return fp
}

func (f FileBuildStorage) Clean() {
	os.RemoveAll(f.saveDirec)
}