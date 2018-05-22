package testutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gorilla/mux"
)

func TarTemplates(t *testing.T, tarLoc, relativeTemplateLoc string) func(t *testing.T) {
	//tar -cvf werker_files.tar *
	here, _ := ioutil.ReadDir(".")
	t.Log(here)
	t.Log("PWD IS", os.ExpandEnv("$PWD"))
	cmdstr := "tar -cvf " + tarLoc +  " *"
	t.Log(cmdstr)
	cmd := exec.Command("/bin/sh", "-c", cmdstr)
	cmd.Dir = relativeTemplateLoc
	var out, err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	errr := cmd.Run()
	t.Log(out.String())
	t.Log(cmd.Dir)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller???? ")
	}
	t.Log(filepath.Dir(filename))
	if errr != nil {
		t.Fatal(fmt.Sprintf("unable to tar up template direc, stdout: %s \n stderr: %s \n err: %s", out.String(), err.String(), errr.Error()))
	}
	return func(t *testing.T) {
		os.Chdir(relativeTemplateLoc)
		errr := os.Remove(tarLoc)
		if errr != nil {
			if !os.IsNotExist(errr) {
				t.Error("couldn't clean up werker_files.tar, error: ", errr.Error())
			}
		}
	}
}


func CreateDoThingsWebServer(tarLoc string, port string) {
	r := mux.NewRouter()
	//_, filename, _, ok := runtime.Caller(0)
	//if !ok {
	//	panic("no caller???? ")
	//}
	r.HandleFunc("/do_things.tar", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, tarLoc)
	})
	http.ListenAndServe(":" + port, r)
}