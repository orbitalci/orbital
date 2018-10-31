package storage

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func NewCabinet(name string) *Cabinet {
	return &Cabinet{name: name}
}

type Cabinet struct {
	files []*Drawer
	name  string
}

type Drawer struct {
	info os.FileInfo
	path string
}

func (f *Cabinet) findLatestFileObj() string {
	var latestFolder = f.files[0]
	for _, folder := range f.files {
		if folder.info.ModTime().After(latestFolder.info.ModTime()) {
			latestFolder = folder
		}
	}
	return latestFolder.path
}

func (f *Cabinet) findFolderPathByName(walkDirec string) (string, error) {
	if err := filepath.Walk(walkDirec, f.folderWalker); err != nil {
		return "", err
	}
	if len(f.files) > 1 {
		var foldernames []string
		for _, fil := range f.files {
			foldernames = append(foldernames, fil.path)
		}
		return "", errors.New(fmt.Sprintf("should not be more than one folder, %d folders were returned: %+v", len(f.files), foldernames))
	}
	if len(f.files) == 0 {
		return "", errors.New("no files found at walkDirec " + walkDirec + " with id " + f.name)
	}
	return f.files[0].path, nil
}

func (f *Cabinet) folderWalker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		if strings.Contains(path, f.name) {
			draw := &Drawer{info, path}
			f.files = append(f.files, draw)
		}
	}
	return nil
}

func (f *Cabinet) fileWalker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if strings.Contains(path, f.name) {
			draw := &Drawer{info, path}
			f.files = append(f.files, draw)
		}
	}
	return nil
}

func (f *Cabinet) isEmpty() bool {
	if len(f.files) == 0 {
		return true
	}
	return false
}

func fileMaker(direc string, id int64, hash string, name string) (string, error) {
	fpath := filepath.Join(direc, hash, strconv.Itoa(int(id)))
	err := os.MkdirAll(fpath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return filepath.Join(fpath, name), nil
}

func getBuildIdFromPath(path string) (int64, error) {
	var buildId int64
	paths := strings.Split(path, string(os.PathSeparator))
	segment := paths[len(paths)-2]
	id, err := strconv.Atoi(segment)
	if err != nil {
		return 0, errors.New("did not find build id in path: " + path + "\nerror: " + err.Error())
	}
	buildId = int64(id)
	//for _, segment := range paths {
	//	id, err := strconv.Atoi(segment)
	//	if err != nil {
	//		continue
	//	}
	//	buildId = int64(id)
	//	break
	//}
	if buildId == 0 {
		return 0, errors.New("did not find build id in path: " + path)
	}
	return buildId, nil
}

// wiill return a subdirec of direc that does not exist. uses random package
func getRandomStorage(direc string) int64 {
	var id int64
	for {
		id = rand.Int63()
		_, err := os.Stat(filepath.Join(direc, strconv.Itoa(int(id))))
		if os.IsNotExist(err) {
			return id
		}
	}
}
