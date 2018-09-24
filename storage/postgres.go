package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

const TimeFormat = "2006-01-02 15:04:05"

var (
	activeRequests = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ocelot_db_active_requests",
		Help: "number of current db requests",
	})
	dbDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ocelot_db_transaction_duration",
		Help:    "database execution times",
		Buckets: prometheus.LinearBuckets(0, 0.25, 15),
		// table: build_summary, etc
		// interaction_type: create | read | update | delete
	}, []string{"table", "interaction_type"})
)

func init() {
	prometheus.MustRegister(activeRequests, dbDuration)
}

func startTransaction() time.Time {
	activeRequests.Inc()
	return time.Now()
}

func finishTransaction(start time.Time, table, crud string) {
	activeRequests.Dec()
	dbDuration.WithLabelValues(table, crud).Observe(time.Since(start).Seconds())
}

func NewPostgresStorage(user string, pw string, loc string, port int, dbLoc string) *PostgresStorage {
	pg := &PostgresStorage{
		user:     user,
		password: pw,
		location: loc,
		port:     port,
		dbLoc:    dbLoc,
	}
	//if err := pg.Connect(); err != nil {
	//	return pg, err
	//}
	return pg
}

type PostgresStorage struct {
	user     string
	password string
	location string
	port     int
	dbLoc    string
	db       *sql.DB
	once     sync.Once
}

func (p *PostgresStorage) Connect() error {
	p.once.Do(func() {
		var err error
		if p.db, err = sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=disable", p.user, p.dbLoc, p.password, p.location, p.port)); err != nil {
			// todo: not sure if we should _kill_ everything.
			ocelog.IncludeErrField(err).Error("couldn't get postgres connection")
			return
		}
		p.db.SetMaxOpenConns(20)
		p.db.SetMaxIdleConns(0)
		p.db.SetConnMaxLifetime(time.Millisecond)
	})
	return nil
}

// todo: need to write a test for this
func (p *PostgresStorage) Healthy() bool {
	err := p.Connect()
	if err != nil {
		return false
	}
	err = p.db.Ping()
	if err != nil {
		ocelog.IncludeErrField(err).Error("ping failed for database")
		return false
	}
	return true
}
func (p *PostgresStorage) Close() {
	err := p.db.Close()
	if err != nil {
		ocelog.IncludeErrField(err).Error("error closing postgres db")
	}
}

//err := p.db.Close()
//if err != nil {
//	ocelog.IncludeErrField(err).Error("error closing")
//}
//}
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

func convertTimestampToTime(stamp *timestamp.Timestamp) time.Time {
	return time.Unix(stamp.Seconds, int64(stamp.Nanos))
}

func convertTimeToTimestamp(tyme time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{Seconds: tyme.Unix()}
}

// AddSumStart updates the build_summary table with the initial information that you get from a webhook
// returning the build id that postgres generates
func (p *PostgresStorage) AddSumStart(hash string, account string, repo string, branch string) (int64, error) {
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "create")
	if err := p.Connect(); err != nil {
		return 0, errors.New("could not connect to postgres: " + err.Error())
	}
	var id int64
	query := `INSERT INTO build_summary(hash, account, repo, branch, status) values ($1,$2,$3,$4,$5) RETURNING id`
	stmt, err := p.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	if err := stmt.QueryRow(hash, account, repo, branch, pb.BuildStatus_NIL).Scan(&id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return id, err
	}
	return id, nil
}

// SetQueueTime will update the QueueTime in the database to the current time
func (p *PostgresStorage) SetQueueTime(id int64) error {
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "update")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `UPDATE build_summary SET queuetime=$1, status=$2 WHERE id=$3`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(time.Now().Format(TimeFormat), pb.BuildStatus_QUEUED, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return nil
}

//StoreFailedValidation will update the rest of the summary fields (failed:true, duration:0)
func (p *PostgresStorage) StoreFailedValidation(id int64) error {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := `UPDATE build_summary SET failed=$1, buildtime=$2, status=$3 WHERE id=$4`
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(true, 0, pb.BuildStatus_FAILED_PRESTART, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return err
}

func (p *PostgresStorage) setStartTime(id int64, stime time.Time) error {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `UPDATE build_summary SET starttime=$1, status=$2 WHERE id=$3`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(stime.Format(TimeFormat), pb.BuildStatus_RUNNING, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return nil
}

func (p *PostgresStorage) StartBuild(id int64) error {
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "update")
	return p.setStartTime(id, time.Now())
}

// UpdateSum updates the remaining fields in the build_summary table
func (p *PostgresStorage) UpdateSum(failed bool, duration float64, id int64) error {
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "update")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	buildstat := pb.BuildStatus_PASSED
	if failed {
		buildstat = pb.BuildStatus_FAILED
	}
	querystr := `UPDATE build_summary SET failed=$1, buildtime=$2, status=$3 WHERE id=$4`
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(failed, duration, buildstat, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return nil
}

func (p *PostgresStorage) RetrieveSum(gitHash string) ([]*pb.BuildSummary, error) {
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "read")
	var sums []*pb.BuildSummary
	if err := p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	var queuetime, starttime time.Time
	querystr := `SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE hash = $1`
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(gitHash)
	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := pb.BuildSummary{}
		err = rows.Scan(&sum.Hash, &sum.Failed, &starttime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch, &queuetime, &sum.Status)
		//fmt.Println(hi)
		if err != nil {
			if err == sql.ErrNoRows {
				return sums, BuildSumNotFound(gitHash)
			}
			ocelog.IncludeErrField(err).Error("failed to retrieve build summary")
			return sums, err
		}
		sum.QueueTime = convertTimeToTimestamp(queuetime)
		sum.BuildTime = convertTimeToTimestamp(starttime)
		sums = append(sums, &sum)
	}
	return sums, nil
}

//RetrieveHashStartsWith will return a list of all hashes starting with the partial string in db
//** THIS WILL ONLY RETURN HASH, ACCOUNT, AND REPO **
func (p *PostgresStorage) RetrieveHashStartsWith(partialGitHash string) ([]*pb.BuildSummary, error) {
	var hashes []*pb.BuildSummary
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "read")
	if err := p.Connect(); err != nil {
		return hashes, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `select distinct (hash), account, repo from build_summary where hash ilike $1`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(partialGitHash + "%")
	if err != nil {
		ocelog.IncludeErrField(err)
		return hashes, err
	}
	defer rows.Close()
	for rows.Next() {
		var result pb.BuildSummary
		err = rows.Scan(&result.Hash, &result.Account, &result.Repo)
		if err != nil {
			if err == sql.ErrNoRows {
				return hashes, BuildSumNotFound(partialGitHash)
			}
			return hashes, err
		}
		hashes = append(hashes, &result)
	}
	return hashes, nil
}

// RetrieveLatestSum will return the latest entry of build_summary where hash starts with `gitHash`
func (p *PostgresStorage) RetrieveLatestSum(partialGitHash string) (*pb.BuildSummary, error) {
	var sum pb.BuildSummary
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "read")
	if err := p.Connect(); err != nil {
		return &sum, errors.New("could not connect to postgres: " + err.Error())
	}
	var queuetime, starttime time.Time
	querystr := `SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE hash ilike $1 ORDER BY id DESC LIMIT 1;`
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return &sum, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(partialGitHash+"%").Scan(&sum.Hash, &sum.Failed, &starttime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch, &queuetime, &sum.Status)
	if err == sql.ErrNoRows {
		ocelog.IncludeErrField(err)
		return &sum, BuildSumNotFound(partialGitHash)
	}
	sum.BuildTime = &timestamp.Timestamp{Seconds: starttime.Unix()}
	sum.QueueTime = &timestamp.Timestamp{Seconds: queuetime.Unix()}
	return &sum, err
}

// RetrieveSumByBuildId will return a build summary based on build id
func (p *PostgresStorage) RetrieveSumByBuildId(buildId int64) (*pb.BuildSummary, error) {
	var sum pb.BuildSummary
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "read")
	if err := p.Connect(); err != nil {
		return &sum, errors.New("could not connect to postgres: " + err.Error())
	}
	var queuetime, starttime time.Time
	querystr := `SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE id = $1`
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return &sum, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(buildId).Scan(&sum.Hash, &sum.Failed, &starttime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch, &queuetime, &sum.Status)
	if err == sql.ErrNoRows {
		ocelog.IncludeErrField(err)
		return &sum, BuildSumNotFound(string(buildId))
	}
	sum.BuildTime = &timestamp.Timestamp{Seconds: starttime.Unix()}
	sum.QueueTime = &timestamp.Timestamp{Seconds: queuetime.Unix()}
	return &sum, err
}

// RetrieveLastFewSums will return < limit> number of summaries that correlate with a repo and account.
func (p *PostgresStorage) RetrieveLastFewSums(repo string, account string, limit int32) ([]*pb.BuildSummary, error) {
	var sums []*pb.BuildSummary
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "read")
	if err := p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := fmt.Sprintf(`SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE repo=$1 and account=$2 ORDER BY id DESC LIMIT %d`, limit)
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var queuetime, starttime time.Time
	rows, err := stmt.Query(repo, account)

	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := pb.BuildSummary{}
		if err = rows.Scan(&sum.Hash, &sum.Failed, &starttime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch, &queuetime, &sum.Status); err != nil {
			if err == sql.ErrNoRows {
				return sums, BuildSumNotFound("account: " + account + "and repo: " + repo)
			}
			return sums, err
		}
		sum.BuildTime = &timestamp.Timestamp{Seconds: starttime.Unix()}
		sum.QueueTime = &timestamp.Timestamp{Seconds: queuetime.Unix()}
		sums = append(sums, &sum)
	}
	return sums, nil
}

// RetrieveAcctRepo will return to you a list of accountname + repositories that matches starting with partialRepo
func (p *PostgresStorage) RetrieveAcctRepo(partialRepo string) ([]*pb.BuildSummary, error) {
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "read")
	var sums []*pb.BuildSummary
	if err := p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := fmt.Sprintf(`select distinct on (account, repo) account, repo from build_summary where repo ilike $1;`)
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(partialRepo + "%")
	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := pb.BuildSummary{}
		if err = rows.Scan(&sum.Account, &sum.Repo); err != nil {
			if err == sql.ErrNoRows {
				return sums, BuildSumNotFound("repository starting with" + partialRepo)
			}
			ocelog.IncludeErrField(err)
			return sums, err
		}
		sums = append(sums, &sum)
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
	start := startTransaction()
	defer finishTransaction(start, "build_output", "create")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	if err := output.Validate(); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	querystr := `INSERT INTO build_output(build_id, output) values ($1,$2)`
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	//"2006-01-02 15:04:05"
	if _, err := stmt.Exec(output.BuildId, output.Output); err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) RetrieveOut(buildId int64) (models.BuildOutput, error) {
	start := startTransaction()
	defer finishTransaction(start, "build_output", "read")
	out := models.BuildOutput{}
	if err := p.Connect(); err != nil {
		return out, errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := `SELECT * FROM build_output WHERE build_id=$1`
	stmt, err := p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return out, err
	}
	defer stmt.Close()
	if err := stmt.QueryRow(buildId).Scan(&out.BuildId, &out.Output, &out.OutputId); err != nil {
		ocelog.IncludeErrField(err)
		return out, err
	}
	return out, nil
}

// RetrieveLastOutByHash will return the last output text that correlates with the gitHash
func (p *PostgresStorage) RetrieveLastOutByHash(gitHash string) (models.BuildOutput, error) {
	start := startTransaction()
	defer finishTransaction(start, "build_output", "read")
	queryStr := "select build_id, output, build_output.id from build_output " +
		"join build_summary on build_output.build_id = build_summary.id and build_summary.hash = $1 " +
		"order by build_summary.queuetime desc limit 1;"
	out := models.BuildOutput{}
	if err := p.Connect(); err != nil {
		return out, errors.New("could not connect to postgres: " + err.Error())
	}
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return out, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(gitHash).Scan(&out.BuildId, &out.Output, &out.OutputId)
	return out, err
}

// AddStageDetail will store the stage data along with a starttime and duration to db
//  The fields required on stageResult to insert into stage detail table are:
// 		stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime, stageResult.StageDuration, stageResult.Status, stageResult.Messages
func (p *PostgresStorage) AddStageDetail(stageResult *models.StageResult) error {
	start := startTransaction()
	defer finishTransaction(start, "build_stage_details", "create")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
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

	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime.Format(TimeFormat), stageResult.StageDuration, stageResult.Status, string(jsonStr)); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}

	return nil
}

// Retrieve StageDetail will return all stages matching build id
func (p *PostgresStorage) RetrieveStageDetail(buildId int64) ([]models.StageResult, error) {
	start := startTransaction()
	defer finishTransaction(start, "build_stage_details", "read")
	var stages []models.StageResult
	queryStr := "select id, build_id, error, starttime, runtime, status, messages, stage from build_stage_details where build_id = $1 order by build_id asc;"
	if err := p.Connect(); err != nil {
		return stages, errors.New("could not connect to postgres: " + err.Error())
	}
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(buildId)
	defer rows.Close()
	for rows.Next() {
		stage := models.StageResult{}
		var errString sql.NullString        //using sql's NullString because calling .Scan
		var messages models.JsonStringArray //have to use custom class because messages are stored in json format

		if err = rows.Scan(&stage.StageResultId, &stage.BuildId, &errString, &stage.StartTime, &stage.StageDuration, &stage.Status, &messages, &stage.Stage); err != nil {
			if err == sql.ErrNoRows {
				return stages, StagesNotFound(fmt.Sprintf("build id: %v", buildId))
			}
			ocelog.IncludeErrField(err).Error()
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

func (p *PostgresStorage) InsertPoll(account string, repo string, cronString string, branches string) (err error) {
	start := startTransaction()
	defer finishTransaction(start, "poll_table", "create")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `INSERT INTO polling_repos(account, repo, cron_string, branches, last_cron_time) values ($1, $2, $3, $4, $5)`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(account, repo, cronString, branches, time.Now()); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).WithField("cronString", cronString).Error("could not insert poll entry into database")
		return
	}
	return
}

func (p *PostgresStorage) UpdatePoll(account string, repo string, cronString string, branches string) (err error) {
	start := startTransaction()
	defer finishTransaction(start, "poll_table", "update")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `UPDATE polling_repos SET (cron_string, branches) = ($1,$2) WHERE (account,repo) = ($3,$4);`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(cronString, branches, account, repo); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).WithField("cronString", cronString).Error("could not update poll entry in database")
		return
	}
	return
}

func (p *PostgresStorage) SetLastData(account string, repo string, lasthashes map[string]string) (err error) {
	start := startTransaction()
	defer finishTransaction(start, "poll_table", "update")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	//starttime.Format(TimeFormat)
	hashes, err := json.Marshal(lasthashes)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't marshal hash map to byte??")
		return err
	}
	queryStr := `UPDATE polling_repos SET (last_cron_time, last_hashes)=($1,$2) WHERE (account,repo) = ($3,$4);`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(time.Now().Format(TimeFormat), hashes, account, repo); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).Error("could not update last_cron_time in database")
	}
	return
}

func (p *PostgresStorage) GetLastData(accountRepo string) (timestamp time.Time, hashes map[string]string, err error) {
	start := startTransaction()
	defer finishTransaction(start, "poll_table", "read")
	if err := p.Connect(); err != nil {

		return time.Now(), nil, errors.New("could not connect to postgres: " + err.Error())
	}
	acctRepo := strings.Split(accountRepo, "/")
	if len(acctRepo) != 2 {
		return time.Now(), nil, errors.New("length on acct repo not correct")
	}
	account, repo := acctRepo[0], acctRepo[1]
	queryStr := `SELECT last_cron_time, last_hashes FROM polling_repos WHERE (account,repo) = ($1,$2);`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return time.Unix(0, 0), nil, err
	}
	defer stmt.Close()
	var bits []byte
	hashes = make(map[string]string)
	if err = stmt.QueryRow(account, repo).Scan(&timestamp, &bits); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("no rows found for " + account + "/" + repo)
			ocelog.IncludeErrField(err).Error("cannot get last cron time or last hashes")
			return timestamp, nil, err
		}
		ocelog.IncludeErrField(err).Error("unable to get last cron time")
		return timestamp, nil, err
	}
	if err = json.Unmarshal(bits, &hashes); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't unmarshal bits")
		return
	}

	ocelog.Log().Debug("returning no errors, everything is TOTALLY FINE")
	return timestamp, hashes, nil
}

func (p *PostgresStorage) PollExists(account string, repo string) (bool, error) {
	start := startTransaction()
	defer finishTransaction(start, "poll_table", "read")
	if err := p.Connect(); err != nil {
		return false, errors.New("could not connect to postgres: " + err.Error())
	}
	var count int64
	queryStr := `select count(*) from polling_repos where (account,repo) = ($1,$2);`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return false, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(account, repo).Scan(&count)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (p *PostgresStorage) DeletePoll(account string, repo string) error {
	start := startTransaction()
	defer finishTransaction(start, "poll_table", "delete")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `delete from polling_repos where (account, repo) =($1,$2)`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(account, repo); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).Error("could not delete poll entry from database")
		return err
	}
	return nil
}

func (p *PostgresStorage) GetAllPolls() ([]*models.PollRequest, error) {
	start := startTransaction()
	defer finishTransaction(start, "poll_table", "read")
	var polls []*models.PollRequest
	if err := p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `select account, repo, cron_string, last_cron_time, branches from polling_repos`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		pr := &models.PollRequest{}
		if err = rows.Scan(&pr.Account, &pr.Repo, &pr.Cron, &pr.LastCron, &pr.Branches); err != nil {
			return nil, err
		}
		polls = append(polls, pr)
	}
	return polls, rows.Err()
}

//type CredTable interface {
//	InsertCred(credder pb.OcyCredder) error
//	// retrieve ordered by cred type
//	RetrieveAllCreds(hideSecret bool) ([]pb.OcyCredder, error)
//	RetrieveCreds(credType pb.CredType, hideSecret bool) ([]pb.OcyCredder, error)
//	RetrieveCred(credType pb.CredType, subCredType pb.SubCredType, accountName string) (pb.OcyCredder, error)
//	HealthyChkr
//}
//CREATE TABLE credentials (
//  id SERIAL PRIMARY KEY,
//  account character varying(100),
//  identifier character varying(100),
//  cred_type smallint,e zone,
//  cred_sub_type smallint,
//  additional_fields jsonb
//);
//

//InsertCred will insert an ocyCredder object into the credentials table after calling its ValidateForInsert method.
// if the OcyCredder fails validation, it will return a *models.ValidationErr
func (p *PostgresStorage) InsertCred(credder pb.OcyCredder, overwriteOk bool) error {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "create")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	if invalid := credder.ValidateForInsert(); invalid != nil {
		return invalid
	}
	//possibleCred, err := p.RetrieveCred(credder.GetSubType(), identifier, accountName string)
	moreFields, err := credder.CreateAdditionalFields()
	if err != nil {
		return errors.New("could not create additional_fields column, error: " + err.Error())
	}
	queryStr := `INSERT INTO credentials(account, identifier, cred_type, cred_sub_type, additional_fields) values ($1,$2,$3,$4,$5)`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType().Parent(), credder.GetSubType(), moreFields)
	return err
}

func (p *PostgresStorage) UpdateCred(credder pb.OcyCredder) error {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "update")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	if invalid := credder.ValidateForInsert(); invalid != nil {
		return invalid
	}
	moreFields, err := credder.CreateAdditionalFields()
	if err != nil {
		return errors.New("could not create additional_fields column, error: " + err.Error())
	}
	queryStr := `UPDATE credentials SET additional_fields=$1 WHERE (account,identifier,cred_sub_type)=($2,$3,$4)`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(moreFields, credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType())
	return err
}

func (p *PostgresStorage) DeleteCred(credder pb.OcyCredder) error {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "delete")
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `DELETE from credentials where (account,identifier,cred_sub_type)=($1,$2,$3)`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare statement")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType())
	return err

}

func (p *PostgresStorage) CredExists(credder pb.OcyCredder) (bool, error) {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err := p.Connect(); err != nil {
		return false, errors.New("could not connect to postgres: " + err.Error())
	}
	var count int64
	queryStr := `select count(*) from credentials where (account,identifier,cred_sub_type) = ($1,$2,$3);`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return false, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType()).Scan(&count)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (p *PostgresStorage) RetrieveAllCreds() ([]pb.OcyCredder, error) {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err := p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `SELECT account, identifier, cred_type, cred_sub_type, additional_fields from credentials order by cred_type`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []pb.OcyCredder
	for rows.Next() {
		var credType, subCredType int32
		var addtlFields []byte
		var account, identifier string
		err := rows.Scan(&account, &identifier, &credType, &subCredType, &addtlFields)
		if err != nil {
			return nil, err
		}
		ocyCredType := pb.CredType(credType)
		cred := ocyCredType.SpawnCredStruct(account, identifier, pb.SubCredType(subCredType))
		if err := cred.UnmarshalAdditionalFields(addtlFields); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	if rows.Err() == sql.ErrNoRows {
		return nil, CredNotFound("all accounts", "all types")
	}
	if len(creds) == 0 {
		return nil, CredNotFound("all accounts", "all types")
	}
	return creds, rows.Err()
}

func (p *PostgresStorage) RetrieveCreds(credType pb.CredType) ([]pb.OcyCredder, error) {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err := p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `SELECT account, identifier, cred_type, cred_sub_type, additional_fields FROM credentials WHERE cred_type=$1`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(credType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []pb.OcyCredder
	for rows.Next() {
		var credType, subCredType int32
		var addtlFields []byte
		var account, identifier string
		err := rows.Scan(&account, &identifier, &credType, &subCredType, &addtlFields)
		if err != nil {
			return nil, err
		}
		ocyCredType := pb.CredType(credType)
		cred := ocyCredType.SpawnCredStruct(account, identifier, pb.SubCredType(subCredType))
		if cred == nil {
			// shouldn't happen?
			return nil, errors.New("unsupported cred type")
		}
		if err := cred.UnmarshalAdditionalFields(addtlFields); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	if rows.Err() == sql.ErrNoRows {
		return creds, CredNotFound("all accounts", credType.String())
	}
	if len(creds) == 0 {
		return creds, CredNotFound("all accounts", credType.String())
	}
	return creds, rows.Err()
}

func (p *PostgresStorage) RetrieveCred(subCredType pb.SubCredType, identifier, accountName string) (pb.OcyCredder, error) {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err := p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `SELECT additional_fields FROM credentials WHERE (cred_sub_type,identifier,account)=($1,$2,$3)`
	ocelog.Log().Debugf("%d %s %s", subCredType, identifier, accountName)
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var addtlFields []byte
	if err := stmt.QueryRow(subCredType, identifier, accountName).Scan(&addtlFields); err != nil {
		if err == sql.ErrNoRows {
			return nil, CredNotFound(accountName, identifier)
		}
		return nil, err
	}
	credder := subCredType.Parent().SpawnCredStruct(accountName, identifier, subCredType)
	if credder == nil {
		// do we even need this check? wouldn't strict typing never allow this condition?
		return nil, errors.New("credder is nil")
	}
	err = credder.UnmarshalAdditionalFields(addtlFields)
	return credder, err
}

func (p *PostgresStorage) RetrieveCredBySubTypeAndAcct(scredType pb.SubCredType, acctName string) ([]pb.OcyCredder, error) {
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err := p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `SELECT additional_fields, identifier FROM credentials WHERE (cred_sub_type,account)=($1,$2)`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(scredType, acctName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []pb.OcyCredder
	for rows.Next() {
		var addtlFields []byte
		var identifier string
		err = rows.Scan(&addtlFields, &identifier)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, CredNotFound(acctName, scredType.String())
			}
			return nil, err
		}
		credder := scredType.Parent().SpawnCredStruct(acctName, identifier, scredType)
		if err = credder.UnmarshalAdditionalFields(addtlFields); err != nil {
			return nil, err
		}
		creds = append(creds, credder)
	}
	if rows.Err() == sql.ErrNoRows {
		return nil, CredNotFound(acctName, scredType.String())
	}
	if len(creds) == 0 {
		return nil, CredNotFound(acctName, scredType.String())
	}
	return creds, rows.Err()
}

func (p *PostgresStorage) GetTrackedRepos() (*pb.AcctRepos, error) {
	start := startTransaction()
	defer finishTransaction(start, "build_summary", "read")
	if err := p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	var queuetime time.Time
	queryStr := `SELECT DISTINCT ON (account, repo) account, repo, queuetime
FROM build_summary
ORDER BY account, repo, queuetime DESC NULLS LAST;`
	stmt, err := p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	acctRepos := &pb.AcctRepos{}
	for rows.Next() {
		acctRepo := &pb.AcctRepo{}
		if err := rows.Scan(&acctRepo.Account, &acctRepo.Repo, &queuetime); err != nil {
			if err == sql.ErrNoRows {
				return nil, BuildSumNotFound("any account/repo")
			}
			return nil, err
		}
		acctRepo.LastQueue = convertTimeToTimestamp(queuetime)
		acctRepos.AcctRepos = append(acctRepos.AcctRepos, acctRepo)
	}
	return acctRepos, nil
}

func (p *PostgresStorage) StorageType() string {
	return fmt.Sprintf("Postgres Database at %s", p.location)
}
