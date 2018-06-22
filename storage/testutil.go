package storage

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/shankj3/ocelot/models/pb"
	"os"
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
func insertDependentData(t *testing.T, pg *PostgresStorage) (int64) {
	hash := "123"
	model := &pb.BuildSummary{
		Hash:          hash,
		Failed:        false,
		BuildTime:     &timestamp.Timestamp{Seconds:time.Now().Unix()},
		Account:       "testAccount",
		BuildDuration: 23.232,
		Repo:          "testRepo",
		Branch:        "aBranch",
	}
	id, err := pg.AddSumStart(model.Hash, model.Account, model.Repo, model.Branch)
	if err != nil {
		t.Error(err)
	}
	return id
}

func createOrUpdateAuditFile(msg string) error {
	f, err := os.OpenFile("./test-fixtures/pg_audit", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.WriteString(msg + "\n"); err != nil {
		return err
	}
	return nil
}

// todo: add some auditing to this, have a file that says what test started it up and what test called cleanup, i know something's missing
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
	del := exec.Command("/bin/sh", "-c", "docker stop pgtest; docker rm pgtest")
	del.Run()
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker run -p %d:5432  -v %s:/docker-entrypoint-initdb.d -e POSTGRES_PASSWORD=%s --name pgtest -d postgres", port, path, password))
	//if err := createOrUpdateAuditFile(fmt.Sprintf("%s,create", t.Name())); err != nil {
	//	t.Error(err)
	//}
	var outbe, errbe bytes.Buffer
	cmd.Stdout = &outbe
	cmd.Stderr = &errbe
	if err := cmd.Run(); err != nil {
		t.Log(outbe.String())
		t.Log(errbe.String())
		t.Fatal("could not start db, err: ", err.Error())
	}
	//var containerId string
	//containerId = strings.Trim(outbe.String(), "\n")
	t.Log("successfully started up test pg database on port 5555")
	cleanup = func(t *testing.T) {
		//createOrUpdateAuditFile(fmt.Sprintf("%s,delete", t.Name()))
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
		cmd = exec.Command("/bin/sh", "-c", "docker kill pgtest")
		if err := cmd.Run(); err != nil {
			t.Log("could not kill db, err: ", err.Error())
		}
		//cmd = exec.Command("/bin/sh", "-c", "docker rm pgtest")
		//if err := cmd.Run(); err != nil {
		//	t.Log("could not rm db image, err: ", err.Error())
		//}
	}
	// this has to happen for postgres to be able to actually start up
	t.Log("waiting 4 seconds for postgres container to start up proper")
	time.Sleep(6 * time.Second)
	return

}

func PostgresTeardown(t *testing.T, db *sql.DB) {
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
