package storage

import (
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func CreateTestFileSystemStorage(t *testing.T) BuildSum {
	return NewFileBuildStorage("./test-fixtures/storage")
}

// create a test postgres database on port 5555 using the official docker image, create the tables, and insert some
// seed data
func insertDependentData(t *testing.T) (*PostgresStorage, int64, func(t *testing.T)){
	cleanup, pw, port := CreateTestPgDatabase(t)
	pg := NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	hash := "123"
	model := &models.BuildSummary{
		Hash: hash,
		Failed: false,
		BuildTime: time.Now(),
		Account: "testAccount",
		BuildDuration: 23.232,
		Repo: "testRepo",
		Branch: "aBranch",
	}
	id, err := pg.AddSumStart(model.Hash, model.Account, model.Repo, model.Branch)
	if err != nil {
		cleanup(t)
		t.Fatal(err)
	}
	return pg, id, cleanup
}

// create the postgres database using docker image, create tables using sql file in test-fixtures
// returns a cleanup function for closing database, the password, and the port it runs on.
func CreateTestPgDatabase(t *testing.T) (cleanup func(t *testing.T), password string, port int) {
	if testing.Short() {
		t.Skip("run flagged as 'short', skipping test that requires docker setup")
	}
	port = 5555
	password = "mysecretpassword"
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	path := filepath.Join(dir, "test-fixtures")
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker run -p %d:5432  -v %s:/docker-entrypoint-initdb.d -e POSTGRES_PASSWORD=%s --name pgtest -d postgres", port, path, password))
	var outbe, errbe bytes.Buffer
	cmd.Stdout = &outbe
	cmd.Stderr = &errbe
	if err := cmd.Run(); err != nil {
		t.Log(outbe.String())
		t.Log(errbe.String())
		t.Fatal("could not start db, err: ", err.Error())
	}
	t.Log("successfully started up test pg database on port 5555")
	cleanup = func(t *testing.T){
		t.Log("attempting to clean up db")
		cmd := exec.Command("/bin/sh", "-c", "docker ps -a | grep pgtest")
		err := cmd.Start()
		var out, errr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errr
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.Wait(); err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				// means that exit code was not zero, so the ps
				return
			} else {
				t.Fatal(err)
			}
		}
		t.Log(out.String())
		t.Log(errr.String())
		cmd = exec.Command("/bin/sh", "-c", "docker kill pgtest")
		if err := cmd.Run(); err != nil {
			t.Fatal("could not kill db, err: ", err.Error())
		}
		cmd = exec.Command("/bin/sh", "-c", "docker rm pgtest")
		if err := cmd.Run(); err != nil {
			t.Fatal("could not rm db image, err: ", err.Error())
		}
	}
	// this has to happen for postgres to be able to actually start up
	t.Log("waiting 4 seconds for postgres container to start up proper")
	time.Sleep(4 * time.Second)
	return

}

func PostgresTeardown(t *testing.T, db *sql.DB) {
	t.Log(db.Stats())
	open := db.Stats().OpenConnections
	if open > 0 {
		t.Fatalf("failed to close %d connections", open)
	}
}

//type HealthyChkr interface {
//	Healthy() bool
//}

func NewHealthyStorage() *HealthyStorage {
	return &HealthyStorage{true}
}

type HealthyStorage struct {
	IsHealthy bool
}

func (h *HealthyStorage) Healthy() bool {
	return h.IsHealthy
}