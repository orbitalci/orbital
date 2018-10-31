package storage

import (
	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/shankj3/go-til/test"
	util "github.com/shankj3/ocelot/common/testutil"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"

	"bytes"
	"testing"
	"time"
)

// run all the tests, so we don't have to start up a bunch of postgress's
func Test_PostgresStorage(t *testing.T) {
	util.BuildServerHack(t)
	port := 5455
	cleanup, pw := CreateTestPgDatabase(t, port)
	defer cleanup(t)
	pg := NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	pg.Connect()
	defer PostgresTeardown(t, pg.db)
	t.Run("get tracked repos", func(t *testing.T) { postgresStorage_GetTrackedRepos(t, pg) })
	t.Run("add sum start", func(t *testing.T) { postgresStorage_AddSumStart(t, pg) })
	id := insertDependentData(t, pg)
	t.Run("get last data", func(t *testing.T) { postgresStorage_GetLastData(t, pg) })
	t.Run("add out", func(t *testing.T) { postgresStorage_AddOut(t, pg, id) })
	t.Run("add stage detail", func(t *testing.T) { postgresStorage_AddStageDetail(t, pg, id) })
	t.Run("add queue time", func(t *testing.T) { postgresStorage_SetQueueTime(t, pg) })
	t.Run("get last successful build hash", func(t *testing.T) { postgresStorage_GetLastSuccessfulBuildHash(t, pg) })
	t.Run("store failed validation", func(t *testing.T) { postgresStorage_StoreFailedValidation(t, pg) })
	t.Run("retrieve hash partial", func(t *testing.T) { postgresStorage_RetrieveHashStartsWith(t, pg) })
	t.Run("cred add", func(t *testing.T) { postgresStorage_InsertCred(t, pg) })
	t.Run("cred delete", func(t *testing.T) { postgresStorage_DeleteCred(t, pg) })
	t.Run("healthy check", func(t *testing.T) { postgresStorage_Healthy(t, pg, cleanup) })
}

func postgresStorage_AddSumStart(t *testing.T, pg *PostgresStorage) {
	const shortForm = "2006-01-02 15:04:05"
	buildTime, err := time.Parse(shortForm, "2018-01-14 18:38:59")
	if err != nil {
		t.Error(err)
	}
	model := &pb.BuildSummary{
		Hash:          "123",
		Failed:        false,
		Status:        pb.BuildStatus_QUEUED,
		BuildTime:     &timestamp.Timestamp{Seconds: buildTime.Unix()},
		Account:       "testAccount",
		BuildDuration: 23.232,
		Repo:          "testRepo",
		Branch:        "aBranch",
	}
	id, err := pg.AddSumStart(model.Hash, model.Account, model.Repo, model.Branch)
	if err != nil {
		t.Error(err)
	}
	t.Log("id ", id)
	sumaries, err := pg.RetrieveSum("123")
	if err != nil {
		t.Error(err)
		return
	}
	sum := sumaries[0]
	if sum.Hash != "123" {
		t.Error(test.StrFormatErrors("hash", "123", sum.Hash))
	}
	// when first inserted, should be true
	if sum.Failed != true {
		t.Error(test.GenericStrFormatErrors("failed", true, sum.Failed))
	}
	//if sum.BuildTime != buildTime {
	//	t.Error(test.GenericStrFormatErrors("build start time", buildTime, sum.BuildTime))
	//}
	if sum.Account != "testAccount" {
		t.Error(test.StrFormatErrors("account", "testAccount", sum.Account))
	}
	if sum.Repo != "testRepo" {
		t.Error(test.StrFormatErrors("repo", "testRepo", sum.Repo))
	}
	if sum.Branch != "aBranch" {
		t.Error(test.StrFormatErrors("branch", "aBranch", sum.Branch))
	}
	err = pg.UpdateSum(model.Failed, model.BuildDuration, id)
	if err != nil {
		t.Error("could not update build summary: ", err)
	}
	model.Status = pb.BuildStatus_PASSED
	//cleanup
	//_ = pg.db.QueryRow(`delete from build_summary where hash = 123`)
	sumaz, err := pg.RetrieveSum("123")
	if err != nil {
		t.Error(err)
	}
	suum := sumaz[0]
	if suum.BuildDuration != model.BuildDuration {
		t.Error(test.GenericStrFormatErrors("build duration", model.BuildDuration, suum.BuildDuration))
	}
	if suum.Failed != false {
		t.Error(test.GenericStrFormatErrors("failed", false, suum.Failed))
	}
	if suum.Status != pb.BuildStatus_PASSED {
		t.Error(test.GenericStrFormatErrors("status", pb.BuildStatus_PASSED, suum.Status))
	}
}

func postgresStorage_AddOut(t *testing.T, pg *PostgresStorage, id int64) {
	txt := []byte("a;lsdkfjakl;sdjfakl;sdjfkl;asdj c389uro23ijrh8234¬˚å˙∆ßˆˆ…∂´¨¨;lsjkdafal;skdur23;klmnvxzic78r39q;lkmsndf")
	out := &models.BuildOutput{
		BuildId: id,
		Output:  txt,
	}
	err := pg.AddOut(out)
	if err != nil {
		t.Error("could not add out: ", err)
	}
	retrieved, err := pg.RetrieveOut(id)
	if err != nil {
		t.Error("could not retrieve out: ", err)
	}
	if retrieved.BuildId != id {
		t.Error(test.GenericStrFormatErrors("build id", id, retrieved.BuildId))
	}
	if bytes.Compare(retrieved.Output, txt) != 0 {
		t.Error(test.GenericStrFormatErrors("output", txt, retrieved.Output))
	}
}

func postgresStorage_AddStageDetail(t *testing.T, pg *PostgresStorage, id int64) {
	const shortForm = "2006-01-02 15:04:05"
	startTime, _ := time.Parse(shortForm, "2018-01-14 18:38:59")
	stageMessage := []string{"wow I am amazing"}

	stageResult := &models.StageResult{
		BuildId:       id,
		Error:         "",
		StartTime:     startTime,
		StageDuration: 100,
		Status:        1,
		Messages:      stageMessage,
		Stage:         "marianne",
	}
	err := pg.AddStageDetail(stageResult)
	if err != nil {
		t.Error("could not add stage details", err)
	}

	stageResults, err := pg.RetrieveStageDetail(id)
	if err != nil {
		t.Error("could not get stage details", err)
	}

	if len(stageResults) != 1 {
		t.Error(test.GenericStrFormatErrors("stage length", 1, len(stageResults)))
	}

	for _, stage := range stageResults {
		if stage.StageResultId != 1 {
			t.Error(test.GenericStrFormatErrors("postgres assigned stage result id", 1, stage.StageResultId))
		}
		if stage.BuildId != id {
			t.Error(test.GenericStrFormatErrors("test build id", 2, stage.BuildId))
		}
		if len(stage.Error) != 0 {
			t.Error(test.GenericStrFormatErrors("stage err length", 0, len(stage.Error)))
		}
		if stage.Stage != "marianne" {
			t.Error(test.GenericStrFormatErrors("stage name", "marianne", stage.Stage))
		}
		if len(stage.Messages) != len(stageMessage) || stage.Messages[0] != stageMessage[0] {
			t.Error(test.GenericStrFormatErrors("stage messages", stageMessage, stage.Messages))
		}
		if stage.StageDuration != 100 {
			t.Error(test.GenericStrFormatErrors("stage duration", 100, stage.Messages))
		}
	}
}

func postgresStorage_Healthy(t *testing.T, pg *PostgresStorage, cleanup func(t2 *testing.T)) {
	if !pg.Healthy() {
		t.Error("postgres storage instance should return healthy, it isn't.")
	}
	cleanup(t)
	time.Sleep(2 * time.Second)
	if pg.Healthy() {
		t.Error("postgres storage instance has been shut down, should return not healthy")
	}
}

func postgresStorage_GetLastData(t *testing.T, pg *PostgresStorage) {
	_, hashes, err := pg.GetLastData("level11consulting/ocelot")
	if err != nil {
		t.Error(err)
	}
	if last, ok := hashes["master"]; !ok {
		t.Error("hash map should have master branch, it doesnlt")
		t.Log(hashes)
	} else if last != "6363a8a4ef13227218dc5c6d40e78ddfeb21b623" {
		t.Error(test.StrFormatErrors("master last hash", "6363a8a4ef13227218dc5c6d40e78ddfeb21b623", last))
	}
}

func postgresStorage_SetQueueTime(t *testing.T, pg *PostgresStorage) {
	id, err := pg.AddSumStart("123", "account", "repo", "master")
	if err != nil {
		t.Error(err)
	}
	err = pg.SetQueueTime(id)
	if err != nil {
		t.Error(err)
	}
	summ, err := pg.RetrieveSumByBuildId(id)
	if err != nil {
		t.Error(err)
	}
	if summ.Status != pb.BuildStatus_QUEUED {
		t.Error(test.GenericStrFormatErrors("status", pb.BuildStatus_QUEUED, summ.Status))
	}
}

func postgresStorage_StoreFailedValidation(t *testing.T, pg *PostgresStorage) {
	id, err := pg.AddSumStart("123", "account", "repo", "master")
	if err != nil {
		t.Error(err)
	}
	err = pg.StoreFailedValidation(id)
	if err != nil {
		t.Error(err)
	}
	summ, err := pg.RetrieveSumByBuildId(id)
	if err != nil {
		t.Error(err)
	}
	if summ.Status != pb.BuildStatus_FAILED_PRESTART {
		t.Error(test.GenericStrFormatErrors("status", pb.BuildStatus_FAILED_PRESTART, summ.Status))
	}
	if summ.Failed != true {
		t.Error(test.GenericStrFormatErrors("failed", true, summ.Failed))
	}

}

func postgresStorage_RetrieveHashStartsWith(t *testing.T, pg *PostgresStorage) {
	_, err := pg.AddSumStart("abcbananahammock", "account", "repo", "master")
	if err != nil {
		t.Error(err)
	}
	_, err = pg.AddSumStart("abcHonkeyTonkey", "account", "repo", "master")
	if err != nil {
		t.Error(err)
	}
	summs, err := pg.RetrieveHashStartsWith("abc")
	if err != nil {
		t.Error(err)
	}
	if len(summs) != 2 {
		t.Errorf("should return both summaries, instead returned %d", len(summs))
	}
	for _, sum := range summs {
		if sum.Hash != "abcbananahammock" && sum.Hash != "abcHonkeyTonkey" {
			t.Error("unexpected build summary")
		}
	}
}

func postgresStorage_GetTrackedRepos(t *testing.T, pg *PostgresStorage) {
	id, err := pg.AddSumStart("hash", "account", "repo", "branch")
	if err != nil { t.Error(err) }
	if err = pg.SetQueueTime(id); err != nil { t.Error(err) }

	id, err = pg.AddSumStart("ha1sh", "account", "repo", "branch1")
	if err != nil {
		t.Error(err)
	}
	if err = pg.SetQueueTime(id); err != nil { t.Error(err) }
	time.Sleep(1)
	id, err = pg.AddSumStart("hash2", "account1", "repo", "branch")
	if err != nil {
		t.Error(err)
	}
	if err = pg.SetQueueTime(id); err != nil { t.Error(err) }

	repos, err := pg.GetTrackedRepos()
	if err != nil {
		t.Error(err)
	}
	if len(repos.AcctRepos) != 2 {
		t.Error("should only be two distinct acct repos")
	}
	for _, repo := range repos.AcctRepos {
		if repo.LastQueue == nil {
			t.Error("last queued time should be set!!")
		}
	}

}

// TODO: add more cred tests here, checking validation etc
func postgresStorage_InsertCred(t *testing.T, pg *PostgresStorage) {
	testCred1 := &pb.GenericCreds{
		Identifier:   "THISBEDACRED",
		SubType:      pb.SubCredType_ENV,
		AcctName:     "OCELOTRULES",
		ClientSecret: "thiswontgetinserted",
	}
	if err := pg.InsertCred(testCred1, true); err != nil {
		t.Error(err)
		return
	}
	retrieved, err := pg.RetrieveCred(pb.SubCredType_ENV, "THISBEDACRED", "OCELOTRULES")
	if err != nil {
		t.Error(err)
	}
	// retrieved isn't going to have the clientsecret since its stored in vault.
	testCred1.ClientSecret = ""
	if diff := deep.Equal(testCred1, retrieved); diff != nil {
		t.Error(diff)
	}
}

// this is expected to run after postgresStorage_insertCred
func postgresStorage_DeleteCred(t *testing.T, pg *PostgresStorage) {
	testCred1 := &pb.GenericCreds{
		Identifier:   "THISBEDACRED",
		SubType:      pb.SubCredType_ENV,
		AcctName:     "OCELOTRULES",
		ClientSecret: "thisisthesecret",
	}
	// make sure it actually is there...
	exists, err := pg.CredExists(testCred1)
	if !exists {
		t.Error("test credential has not been inserted.. someone is executing this incorrectly. ")
	}
	if err != nil {
		t.Error("could not connect w/ database, error: " + err.Error())
		return
	}
	if err := pg.DeleteCred(testCred1); err != nil {
		t.Error(err)
	}
	_, err = pg.RetrieveCred(pb.SubCredType_ENV, "THISBEDACRED", "OCELOTRULES")
	if err == nil {
		t.Error("error should not be nil! this should be deleted!")
		return
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Error("should be a not found error? wtf? instead its: ", err.Error())
	}
}


func postgresStorage_GetLastSuccessfulBuildHash(t *testing.T, pg *PostgresStorage) {
	var data = []struct{
		hash string
		branch string
		failed bool
	}{
		{"hash1", "branch1", true},
		{"hash2", "branch1", true},
		{"hash10", "branch1", false},

		{"hash4", "branch2", true},
		{"hash5", "branch2", false},
		{"hash6", "branch2", true},

		{"hash7", "branch3", true},
	}
	for _, datum := range data {
		id, err := pg.AddSumStart(datum.hash, "account", "repo", datum.branch)
		if err != nil {
			t.Error(err)
			return
		}
		pg.SetQueueTime(id)
		time.Sleep(2)
		if err = pg.UpdateSum(datum.failed, 10.0112, id); err != nil {
			t.Error(err)
			return
		}
	}
	lastHash, err := pg.GetLastSuccessfulBuildHash("account", "repo", "branch3")
	if err == nil {
		t.Error("branch3 has no sucessful builds, this should fail")
		return
	}

	lastHash, err = pg.GetLastSuccessfulBuildHash("account", "repo", "branch2")
	if err != nil {
		t.Error(err)
		return
	}
	if lastHash != "hash5" {
		t.Error("branch2 has one successful build, hash5. this returned " + lastHash)
	}
	lastHash, err = pg.GetLastSuccessfulBuildHash("account", "repo", "branch1")
	if err != nil {
		t.Error(err)
		return
	}
	if lastHash != "hash10" {
		t.Error("branch2 has two succesful builds, and the latest is hash10. this returned " + lastHash)
	}

}