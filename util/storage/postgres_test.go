package storage

import (
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"testing"
	"time"
)

func insertDependentData(t *testing.T) (*PostgresStorage, int64){
	pg := NewPostgresStorage("postgres", "mysecretpassword", "localhost", 5432)
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
		t.Fatal(err)
	}
	return pg, id
}

func TestPostgresStorage_AddSumStart(t *testing.T) {
	pg := NewPostgresStorage("postgres", "mysecretpassword", "localhost", 5432)
	const shortForm = "2006-01-02 15:04:05"
	buildTime, err := time.Parse(shortForm,"2018-01-14 18:38:59")
	if err != nil {
		t.Fatal(err)
	}
	model := &models.BuildSummary{
		Hash: "123",
		Failed: false,
		BuildTime: buildTime,
		Account: "testAccount",
		BuildDuration: 23.232,
		Repo: "testRepo",
		Branch: "aBranch",
	}
	id, err := pg.AddSumStart(model.Hash, model.BuildTime, model.Account, model.Repo, model.Branch)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("id ", id)
	sumaries, err := pg.RetrieveSum("123")
	if err != nil {
		t.Fatal(err)
	}
	sum := sumaries[0]
	if sum.Hash != "123" {
		t.Error(test.StrFormatErrors("hash", "123", sum.Hash))
	}
	if sum.Failed != false {
		t.Error(test.GenericStrFormatErrors("failed", false, sum.Failed))
	}
	//if sum.BuildTime != buildTime {
	//	t.Error(test.GenericStrFormatErrors("build start time", buildTime, sum.BuildTime))
	//}
	if sum.Account != "testAccount" {
		t.Error(test.StrFormatErrors("account", "testAccount", sum.Account))
	}
	if sum.BuildDuration != 23.232 {
		t.Error(test.GenericStrFormatErrors("build duration", 23.232, sum.BuildDuration))
	}
	if sum.Repo != "testRepo" {
		t.Error(test.StrFormatErrors("repo", "testRepo", sum.Repo))
	}
	if sum.Branch != "aBranch" {
		t.Error(test.StrFormatErrors("branch", "aBranch", sum.Branch))
	}
	//cleanup
	_ = pg.db.QueryRow(`delete from build_summary where hash = 123`)
}

func TestPostgresStorage_AddOut(t *testing.T) {
	pg, id := insertDependentData(t)
	txt := "a;lsdkfjakl;sdjfakl;sdjfkl;asdj c389uro23ijrh8234¬˚å˙∆ßˆˆ…∂´¨¨;lsjkdafal;skdur23;klmnvxzic78r39q;lkmsndf"
	out := &models.BuildOutput{
		BuildId: id,
		Output: txt,
	}
	err := pg.AddOut(out)
	defer pg.db.QueryRow(`delete from build_summary where hash = 123`)
	defer pg.db.QueryRow(`delete from build_output where build_id = $1`, id)
	if err != nil {
		t.Fatal(err)
	}
	retrieved, err := pg.RetrieveOut(id)
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.BuildId != id {
		t.Error(test.GenericStrFormatErrors("build id", id, retrieved.BuildId))
	}
	if retrieved.Output != txt {
		t.Error(test.StrFormatErrors("output", txt, retrieved.Output))
	}

}

func TestPostgresStorage_AddFail(t *testing.T) {
	pg, id := insertDependentData(t)
	defer pg.db.QueryRow(`delete from build_summary where id = $1`, id)
	adtl := make(models.FailureData)
	adtl["sup"] = "123"
	fails := &models.FailureReasons{
		Stage: "weeeee",
		Status: 0,
		Error: "ayyyyyy it broke mayn",
		Messages: []string{"why u broken????"},
		Additional: adtl,

	}
	bfr := &models.BuildFailureReason{
		BuildId: id,
		FailureReasons: fails,
	}
	err := pg.AddFail(bfr)
	defer pg.db.QueryRow(`delete from build_failure_reason where build_id = $1`, id)
	if err != nil {
		t.Fatal(err)
	}

	retrieved, err := pg.RetrieveFail(id)
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.FailureReasons.Stage != "weeeee" {
		t.Error(test.StrFormatErrors("stage", "weeeee", retrieved.FailureReasons.Stage))
	}
	if retrieved.FailureReasons.Error != "ayyyyyy it broke mayn" {
		t.Error(test.StrFormatErrors("error", "ayyyyyy it broke mayn", retrieved.FailureReasons.Error))
	}
	if retrieved.FailureReasons.Messages[0] != "why u broken????" {
		t.Error(test.StrFormatErrors("first message", "why u broken????", retrieved.FailureReasons.Messages[0]))
	}
	if retrieved.FailureReasons.Additional["sup"] != "123" {
		t.Fail()
	}
	t.Log(retrieved.FailureReasons.Additional)

}