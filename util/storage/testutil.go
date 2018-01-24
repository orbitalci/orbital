package storage

import (
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// create a test postgres database on port 5555 using the official docker image, create the tables, and insert some
// seed data
func insertDependentData(t *testing.T) (*PostgresStorage, int64, func(t *testing.T)){
	cleanup, pw, port := CreateTestPgDatabase(t)
	pg := NewPostgresStorage("postgres", pw, "localhost", port)
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
	id, err := pg.AddSumStart(model.Hash, model.BuildTime, model.Account, model.Repo, model.Branch)
	if err != nil {
		cleanup(t)
		t.Fatal(err)
	}
	return pg, id, cleanup
}

// create the postgres database using docker image, create tables using sql file in test-fixtures
// returns a cleanup function for closing database, the password, and the port it runs on.
func CreateTestPgDatabase(t *testing.T) (cleanup func(t *testing.T), password string, port int) {
	port = 5555
	password = "mysecretpassword"
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker run -p %d:5432 -e POSTGRES_PASSWORD=%s --name pgtest -d postgres", port, password))
	if err := cmd.Run(); err != nil {
		t.Fatal("could not start db, err: ", err.Error())
	}
	t.Log("successfully started up test pg database on port 5555")
	cleanup = func(t *testing.T){
		cmd := exec.Command("/bin/sh", "-c", "docker kill pgtest")
		if err := cmd.Run(); err != nil {
			t.Fatal("could not kill db, err: ", err.Error())
		}
		cmd = exec.Command("/bin/sh", "-c", "docker rm pgtest")
		if err := cmd.Run(); err != nil {
			t.Fatal("could not rm db image, err: ", err.Error())
		}
	}
	path, err := filepath.Abs("./test-fixtures/schema.sql")
	if err != nil {
		t.Fatal("could not get absolute path to schema sql file")
	}
	psqlCmd := fmt.Sprintf(`PGPASSWORD=mysecretpassword psql -h localhost -p 5555 --user postgres < %s`, path)
	cmd = exec.Command("/bin/sh", "-c", psqlCmd)
	if err := cmd.Run(); err != nil {
		t.Fatal("could not insert schemas")
	}
	return

}