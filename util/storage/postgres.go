package storage


import (
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"strings"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

func NewPostgresStorage(user string, pw string, loc string, port int, dbLoc string) *PostgresStorage {
	pg := &PostgresStorage{
		user: user,
		password: pw,
		location: loc,
		port: port,
		dbLoc: dbLoc,
	}
	//if err := pg.Connect(); err != nil {
	//	return pg, err
	//}
	return pg
}

type PostgresStorage struct {
	user string
	password string
	location string
	port int
	dbLoc string
	db *sql.DB
}

func (p *PostgresStorage) Connect() error {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=disable", p.user, p.dbLoc, p.password, p.location, p.port))
	if err != nil {
		ocelog.IncludeErrField(err)
		return err
	}
	p.db = db
	return nil
}

func (p *PostgresStorage) Disconnect() {
	err := p.db.Close()
	if err != nil {
		ocelog.IncludeErrField(err)
	}
}
/*
Column   |            Type             |
-----------+----------------------------
hash      | character varying(50)
failed    | boolean
starttime | timestamp without time zone
account   | character varying(50)
buildtime | numeric
repo      | character varying(100)
id        | integer
branch    | character varying
*/

// AddSumStart updates the build_summary table with the initial information that you get from a webhook
// returning the build id that postgres generates
func (p *PostgresStorage) AddSumStart(hash string, starttime time.Time, account string, repo string, branch string) (int64, error) {
	if err := p.Connect(); err != nil {
		return 0, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	var id int64
	if err := p.db.QueryRow(`INSERT INTO build_summary(hash, starttime, account, repo, branch, failed, buildtime) values ($1,$2,$3,$4,$5,true,-99.999) RETURNING id`,
		hash, starttime.Format(TimeFormat), account, repo, branch).Scan(&id); err != nil {
		ocelog.IncludeErrField(err)
		return id, err
	}
	return id, nil
}

// UpdateSum updates the remaining fields in the build_summary table
func (p *PostgresStorage) UpdateSum(failed bool, duration float64, id int64) error {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	querystr := `UPDATE build_summary SET failed=$1, buildtime=$2 WHERE id=$3`
	if _, err := p.db.Query(querystr, failed, duration, id); err != nil {
		ocelog.IncludeErrField(err)
		return err
	}
	return nil
}

func (p *PostgresStorage) RetrieveSum(gitHash string) ([]models.BuildSummary, error) {
	var sums []models.BuildSummary
	if err := p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	rows, err := p.db.Query(`SELECT * FROM build_summary WHERE hash = $1`, gitHash)
	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := models.BuildSummary{}
		err = rows.Scan(&sum.Hash, &sum.Failed, &sum.BuildTime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch)
		if err != nil {
			if err == sql.ErrNoRows {
				return sums, BuildSumNotFound(gitHash)
			}
			ocelog.IncludeErrField(err)
			return sums, err
		}
		//fmt.Println(hi)
		sums = append(sums, sum)
	}
	return sums, nil
}

//RetrieveHashStartsWith will return a list of all hashes starting with the partial string in db
//** THIS WILL ONLY RETURN HASH, ACCOUNT, AND REPO **
func (p *PostgresStorage) RetrieveHashStartsWith(partialGitHash string) ([]models.BuildSummary, error) {
	var hashes []models.BuildSummary
	if err := p.Connect(); err != nil {
		return hashes, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	rows, err := p.db.Query(`select distinct (hash), account, repo from build_summary where hash ilike $1`, partialGitHash + "%")
	if err != nil {
		ocelog.IncludeErrField(err)
		return hashes, err
	}
	defer rows.Close()
	for rows.Next() {
		var result models.BuildSummary
		err = rows.Scan(&result.Hash, &result.Account, &result.Repo)
		if err != nil {
			if err == sql.ErrNoRows {
				return hashes, BuildSumNotFound(partialGitHash)
			}
			return hashes, err
		}
		hashes = append(hashes, result)
	}
	return hashes, nil
}

// RetrieveLatestSum will return the latest entry of build_summary where hash starts with `gitHash`
func (p *PostgresStorage) RetrieveLatestSum(partialGitHash string) (models.BuildSummary, error) {
	var sum models.BuildSummary
	if err := p.Connect(); err != nil {
		return sum, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	querystr := `SELECT * FROM build_summary WHERE hash ilike $1 ORDER BY starttime DESC LIMIT 1;`
	row := p.db.QueryRow(querystr, partialGitHash + "%")
	err := row.Scan(&sum.Hash, &sum.Failed, &sum.BuildTime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch)
	if err == sql.ErrNoRows {
		ocelog.IncludeErrField(err)
		return sum, BuildSumNotFound(partialGitHash)
	}
	return sum, err
}

// RetrieveSumByBuildId will return a build summary based on build id
func (p *PostgresStorage) RetrieveSumByBuildId(buildId int64) (models.BuildSummary, error) {
	var sum models.BuildSummary
	if err := p.Connect(); err != nil {
		return sum, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	querystr := `SELECT * FROM build_summary WHERE id = $1 ORDER BY starttime DESC LIMIT 1`
	row := p.db.QueryRow(querystr, buildId)
	err := row.Scan(&sum.Hash, &sum.Failed, &sum.BuildTime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch)
	if err == sql.ErrNoRows {
		ocelog.IncludeErrField(err)
		return sum, BuildSumNotFound(string(buildId))
	}
	return sum, err
}

// RetrieveLastFewSums will return <limit> number of summaries that correlate with a repo and account.
func (p *PostgresStorage) RetrieveLastFewSums(repo string, account string, limit int32) ([]models.BuildSummary, error) {
	var sums []models.BuildSummary
	if err := p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	queryRow := fmt.Sprintf(`SELECT * FROM build_summary WHERE repo=$1 and account=$2 ORDER BY starttime DESC LIMIT %d`, limit)
	rows, err := p.db.Query(queryRow, repo, account)
	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := models.BuildSummary{}
		if err = rows.Scan(&sum.Hash, &sum.Failed, &sum.BuildTime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch); err != nil {
			if err == sql.ErrNoRows {
				return sums, BuildSumNotFound("account: " + account + "and repo: " + repo)
			}
			return sums, err
		}
		sums = append(sums, sum)
	}
	return sums, nil
}

// RetrieveAcctRepo will return to you a list of accountname + repositories that matches starting with partialRepo
func (p *PostgresStorage) RetrieveAcctRepo(partialRepo string) ([]models.BuildSummary, error) {
	var sums []models.BuildSummary
	if err := p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	queryRow := fmt.Sprintf(`select distinct on (account, repo) account, repo from build_summary where repo ilike $1;`)
	rows, err := p.db.Query(queryRow, partialRepo + "%")
	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := models.BuildSummary{}
		if err = rows.Scan(&sum.Account, &sum.Repo); err != nil {
			if err == sql.ErrNoRows {
				return sums, BuildSumNotFound("repository starting with" + partialRepo)
			}
			ocelog.IncludeErrField(err)
			return sums, err
		}
		sums = append(sums, sum)
	}
	return sums, nil
}

/*
  Column  |       Type        | Collation | Nullable
----------+-------------------+-----------+-----------
 build_id | integer           |           | not null
 output   | character varying |           |
 id       | integer           |           | not null
 */

 //AddOut adds build output text to build_output table
func (p *PostgresStorage) AddOut(output *models.BuildOutput) error {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	if err := output.Validate(); err != nil {
		ocelog.IncludeErrField(err)
		return err
	}
	queryStr := `INSERT INTO build_output(build_id, output) values ($1,$2)`
	//"2006-01-02 15:04:05"
	if _, err := p.db.Query(queryStr, output.BuildId, output.Output); err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) RetrieveOut(buildId int64) (models.BuildOutput, error) {
	out := models.BuildOutput{}
	if err := p.Connect(); err != nil {
		return out, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	queryStr := `SELECT * FROM build_output WHERE build_id=$1`
	if err := p.db.QueryRow(queryStr, buildId).Scan(&out.BuildId, &out.Output, &out.OutputId); err != nil {
		ocelog.IncludeErrField(err)
		return out, err
	}
	return out, nil
}

// RetrieveLastOutByHash will return the last output text that correlates with the gitHash
func (p *PostgresStorage) RetrieveLastOutByHash(gitHash string) (models.BuildOutput, error) {
	queryStr := "select build_id, output, build_output.id from build_output " +
		"join build_summary on build_output.build_id = build_summary.id and build_summary.hash = $1 " +
			"order by build_summary.starttime desc limit 1;"
	out := models.BuildOutput{}
	if err := p.Connect(); err != nil {
		return out, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	err := p.db.QueryRow(queryStr, gitHash).Scan(&out.BuildId, &out.Output, &out.OutputId)
	return out, err
}


// AddStageDetail will store the stage data along with a starttime and duration to db
//  The fields required on stageResult to insert into stage detail table are:
// 		stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime, stageResult.StageDuration, stageResult.Status, stageResult.Messages

func (p *PostgresStorage) AddStageDetail(stageResult *models.StageResult) error {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	if err := stageResult.Validate(); err != nil {
		ocelog.IncludeErrField(err)
		return err
	}
	queryStr := `INSERT INTO build_stage_details(build_id, stage, error, starttime, runtime, status, messages) values ($1, $2, $3, $4, $5, $6, $7)`
	jsonStr, err := json.Marshal(stageResult.Messages)
	if err != nil {
		ocelog.IncludeErrField(err)
		return err
	}

	var rows *sql.Rows

	if rows, err = p.db.Query(queryStr, stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime.Format(TimeFormat), stageResult.StageDuration, stageResult.Status, string(jsonStr)); err != nil {
		ocelog.IncludeErrField(err)
		return err
	}

	//don't know why at this point it would also be null
	if rows != nil {
		rows.Scan()
	}

	return nil
}

// Retrieve StageDetail will return all stages matching build id
func(p *PostgresStorage) RetrieveStageDetail(buildId int64) ([]models.StageResult, error) {
	var stages []models.StageResult
	queryStr := "select id, build_id, error, starttime, runtime, status, messages, stage from build_stage_details where build_id = $1 order by starttime asc;"
	if err := p.Connect(); err != nil {
		return stages, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	rows, err := p.db.Query(queryStr, buildId)

	defer rows.Close()
	for rows.Next() {
		stage := models.StageResult{}
		var errString sql.NullString //using sql's NullString because calling .Scan
		var messages models.JsonStringArray //have to use custom class because messages are stored in json format

		if err = rows.Scan(&stage.StageResultId, &stage.BuildId, &errString, &stage.StartTime, &stage.StageDuration, &stage.Status, &messages, &stage.Stage); err != nil {
			if err == sql.ErrNoRows {
				return stages, StagesNotFound(fmt.Sprintf("build id: %v", buildId))
			}
			ocelog.IncludeErrField(err)
			return stages, err
		}

		if errString.Valid {
			stage.Error = errString.String
		}
		stage.Messages = messages
		stages = append(stages, stage)
	}
	return stages, err
}


func (p *PostgresStorage) InsertPoll(account string, repo string, activeCron bool, cronString string, branches string) (err error) {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	queryStr := `INSERT INTO polling_repos(account, repo, cron_ed, cron_string, branches, last_cron_time) values ($1, $2, $3, $4, $5, $6)`
	if _, err = p.db.Exec(queryStr, account, repo, activeCron, cronString, branches, time.Now()); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).WithField("cronString", cronString).Error("could not insert poll entry into database")
		return
	}
	return
}

func (p *PostgresStorage) UpdatePoll(account string, repo string, activeCron bool, cronString string, branches string) (err error) {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	queryStr := `UPDATE polling_repos SET (cron_ed, cron_string, branches) = ($1,$2,$3) WHERE (account,repo) = ($4,$5);`
	if _, err = p.db.Exec(queryStr, activeCron, cronString, branches, account, repo); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).WithField("cronString", cronString).Error("could not update poll entry in database")
		return
	}
	return
}

func (p *PostgresStorage) SetLastCronTime(account string, repo string) (err error) {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	//starttime.Format(TimeFormat)
	queryStr := `UPDATE polling_repos SET (last_cron_time) = ROW($1) WHERE (account,repo) = ($2,$3);`
	if _, err := p.db.Exec(queryStr, time.Now().Format(TimeFormat), account, repo); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).Error("could not update last_cron_time in database")
	}
	return
}

func (p *PostgresStorage) GetLastCronTime(accountRepo string) (timestamp time.Time, err error) {
	if err := p.Connect(); err != nil {
		return time.Now(), errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	fmt.Println(p.dbLoc, p.db, p.location)
	acctRepo := strings.Split(accountRepo, "/")
	if len(acctRepo) != 2 {
		return time.Now(), errors.New("length on acct repo not correct")
	}
	account, repo := acctRepo[0], acctRepo[1]
	queryStr := `SELECT last_cron_time FROM polling_repos WHERE (account,repo) = ($1,$2);`
	var row *sql.Row
	row = p.db.QueryRow(queryStr, account, repo)
	if row == nil {
		err = errors.New("no rows found for " + account + "/" + repo)
		ocelog.IncludeErrField(err).Error("cannot get last cron time")
		return timestamp, err
	}
	if err = row.Scan(&timestamp); err != nil {
		ocelog.IncludeErrField(err).Error("unable to get last cron time")
		return timestamp, err
	}
	ocelog.Log().Debug("returning no errors, everything is TOTALLY FINE")
	return timestamp, nil
}

func (p *PostgresStorage) PollExists(account string, repo string) (bool, error) {
	if err := p.Connect(); err != nil {
		return false, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	var count int64
	queryStr := `select count(*) from polling_repos where (account,repo) = ($1,$2);`
	err := p.db.QueryRow(queryStr, account, repo).Scan(&count)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	} else {
		return true, nil
	}
}


func (p *PostgresStorage) GetAllPolls() ([]*models.PollRequest, error) {
	var polls []*models.PollRequest
	if err := p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	queryStr := `select account, repo, cron_string, last_cron_time, branches from polling_repos`
	rows, err := p.db.Query(queryStr)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		pr := &models.PollRequest{}
		if err = rows.Scan(&pr.Account, &pr.Repo, &pr.Cron, &pr.LastCron, &pr.Branches); err != nil {
			return nil, err
		}
		polls = append(polls, pr)
	}
	return polls, rows.Err()
}



func (p *PostgresStorage) StorageType() string {
	return fmt.Sprintf("Postgres Database at %s", p.location)
}
