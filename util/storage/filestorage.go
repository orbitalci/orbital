package storage

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)



// FileBuildStorage is an implementation of BuildOutput that is for filesystem.
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
	if err = os.MkdirAll(f.saveDirec, 0755); err != nil {
		ocelog.IncludeErrField(err).Fatal("could not init build storage")
	}
}

func (f *FileBuildStorage) StorageType() string {
	return "FileSystem Ocelot Storage"
}

// retrieve build data from filesystem
func (f *FileBuildStorage) Retrieve(gitHash string) (data []byte, err error) {
	fp := f.getTempFile(gitHash)
	data, err = ioutil.ReadFile(fp)
	return
}

func (f *FileBuildStorage) getTempFile(gitHash string) string {
	fp := filepath.Join(f.saveDirec, gitHash)
	return fp
}
func (f *FileBuildStorage) Clean() {
	os.RemoveAll(f.saveDirec)
}


func (f *FileBuildStorage) AddSumStart(hash string, starttime time.Time, account string, repo string, branch string) (int64, error) {
	id := getRandomStorage(f.saveDirec)
	file, err := fileMaker(f.saveDirec, id, hash, "sum.json")
	if err != nil {
		return id, err
	}
	sum := &models.BuildSummary{
		Hash: hash,
		BuildTime: starttime,
		Account: account,
		Repo: repo,
		Branch: branch,
	}
	bytez, err := json.Marshal(sum)
	if err != nil {
		// todo.. delete on fail?
		return id, err
	}
	err = ioutil.WriteFile(file, bytez, os.ModePerm)
	return id, err
}

func (f *FileBuildStorage) UpdateSum(failed bool, duration float64, id int64) error {
	cab := NewCabinet(strconv.Itoa(int(id)))
	path, err := cab.findFolderPathByName(f.saveDirec)
	if err != nil {
		return err
	}
	filename := filepath.Join(path, "sum.json")
	if err != nil {
		return err
	}
	bytez, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var oldSum models.BuildSummary
	if err := json.Unmarshal(bytez, &oldSum); err != nil {
		return err
	}
	oldSum.Failed = failed
	oldSum.BuildDuration = duration
	bytez, err = json.Marshal(oldSum)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, bytez, os.ModePerm)
	return err
}

func (f *FileBuildStorage) RetrieveSum(gitHash string) ([]models.BuildSummary, error) {
	var sums []models.BuildSummary
	cab := NewCabinet("sum.json")
	err := filepath.Walk(filepath.Join(f.saveDirec, gitHash), cab.fileWalker)
	if err != nil {
		return sums, err
	}
	if cab.isEmpty() {
		return sums, BuildSumNotFound(gitHash)
	}
	for _, drawer := range cab.files {
		file, err := ioutil.ReadFile(drawer.path)
		if err != nil {
			return sums, err
		}
		var sum models.BuildSummary
		if err := json.Unmarshal(file, &sum); err != nil {
			return sums, err
		}
		sums = append(sums, sum)

	}
	return sums, nil
}

func (f *FileBuildStorage) RetrieveLatestSum(gitHash string) (models.BuildSummary, error) {
	var summary models.BuildSummary
	cab := NewCabinet("sum.json")
	err := filepath.Walk(filepath.Join(f.saveDirec, gitHash), cab.fileWalker)
	if err != nil {
		return summary, err
	}
	if cab.isEmpty() {
		return summary, BuildSumNotFound(gitHash)
	}
	fp := cab.findLatestFileObj()
	bytez, err := ioutil.ReadFile(fp)
	if err != nil {
		return summary, err
	}
	err = json.Unmarshal(bytez, &summary)
	return summary, err
}

func (f *FileBuildStorage) AddOut(output *models.BuildOutput) error {
	if err := output.Validate(); err != nil {
		return err
	}
	cab := NewCabinet(string(output.BuildId))
	fp, err := cab.findFolderPathByName(f.saveDirec)
	if err != nil {
		return err
	}
	filename := filepath.Join(fp, "out.json")
	bytez, err := json.Marshal(output)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, bytez, os.ModePerm)
	return err

}

func (f *FileBuildStorage) RetrieveOut(buildId int64) (models.BuildOutput, error) {
	var out models.BuildOutput
	cab := NewCabinet(string(buildId))
	fp, err := cab.findFolderPathByName(f.saveDirec)
	if err != nil {
		return out, err
	}
	if cab.isEmpty() {
		return out, BuildOutNotFound(strconv.Itoa(int(buildId)))
	}
	bytez, err := ioutil.ReadFile(filepath.Join(fp, "out.json"))
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(bytez, &out)
	return out, err
}

func (f *FileBuildStorage) RetrieveLastOutByHash(gitHash string) (models.BuildOutput, error) {
	var out models.BuildOutput
	cab := NewCabinet(gitHash)
	if err := filepath.Walk(f.saveDirec, cab.folderWalker); err != nil {
		return out, err
	}
	latestDirec := cab.findLatestFileObj()
	bytez, err := ioutil.ReadFile(filepath.Join(latestDirec, "out.json"))
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(bytez, &out)
	return out, err
}

// todo; implement add fail / retrieve fail
func (f *FileBuildStorage) AddFail(reason *models.BuildFailureReason) error {
	return nil
}

func (f *FileBuildStorage) RetrieveFail(buildId int64) (models.BuildFailureReason, error) {
	return models.BuildFailureReason{}, nil
}