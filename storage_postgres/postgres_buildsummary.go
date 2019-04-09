package storage_postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/level11consulting/ocelot/models/pb"
	storage_error "github.com/level11consulting/ocelot/storage/error"
	metrics "github.com/level11consulting/ocelot/storage/metrics"
	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"
)

// AddSumStart updates the build_summary table with the initial information that you get from a webhook
// returning the build id that postgres generates
func (p *PostgresStorage) AddSumStart(hash string, account string, repo string, branch string, by pb.SignaledBy, credentialsId int64) (int64, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "create")
	if err = p.Connect(); err != nil {
		return 0, errors.New("could not connect to postgres: " + err.Error())
	}
	var id int64
	query := `INSERT INTO build_summary(hash, account, repo, branch, status, signaled_by, credentials_id) values ($1,$2,$3,$4,$5, $6, $7) RETURNING id`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	if err = stmt.QueryRow(hash, account, repo, branch, pb.BuildStatus_NIL, by, credentialsId).Scan(&id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return id, err
	}
	return id, nil
}

// SetQueueTime will update the QueueTime in the database to the current time
func (p *PostgresStorage) SetQueueTime(id int64) error {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "update")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `UPDATE build_summary SET queuetime=$1, status=$2 WHERE id=$3`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(time.Now().Format(TimeFormat), pb.BuildStatus_QUEUED, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return nil
}

//StoreFailedValidation will update the rest of the summary fields (failed:true, duration:0)
func (p *PostgresStorage) StoreFailedValidation(id int64) error {
	var err error
	defer metrics.MetricizeDbErr(err)
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := `UPDATE build_summary SET failed=$1, buildtime=$2, status=$3 WHERE id=$4`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(true, 0, pb.BuildStatus_FAILED_PRESTART, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return err
}

func (p *PostgresStorage) setStartTime(id int64, stime time.Time) error {
	var err error
	defer metrics.MetricizeDbErr(err)
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `UPDATE build_summary SET starttime=$1, status=$2 WHERE id=$3`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(stime.Format(TimeFormat), pb.BuildStatus_RUNNING, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return nil
}

func (p *PostgresStorage) StartBuild(id int64) error {
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "update")
	return p.setStartTime(id, time.Now())
}

// UpdateSum updates the remaining fields in the build_summary table
func (p *PostgresStorage) UpdateSum(failed bool, duration float64, id int64) error {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "update")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	buildstat := pb.BuildStatus_PASSED
	if failed {
		buildstat = pb.BuildStatus_FAILED
	}
	querystr := `UPDATE build_summary SET failed=$1, buildtime=$2, status=$3 WHERE id=$4`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(failed, duration, buildstat, id); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	return nil
}

func (p *PostgresStorage) RetrieveSum(gitHash string) (sums []*pb.BuildSummary, err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	if err = p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	var queuetime, starttime time.Time
	querystr := `SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE hash = $1`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(gitHash)
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
				return sums, storage_error.BuildSumNotFound(gitHash)
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
func (p *PostgresStorage) RetrieveHashStartsWith(partialGitHash string) (hashes []*pb.BuildSummary, err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	if err = p.Connect(); err != nil {
		return hashes, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `select distinct (hash), account, repo from build_summary where hash ilike $1`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(partialGitHash + "%")
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
				return hashes, storage_error.BuildSumNotFound(partialGitHash)
			}
			return hashes, err
		}
		hashes = append(hashes, &result)
	}
	return hashes, nil
}

// RetrieveLatestSum will return the latest entry of build_summary where hash starts with `gitHash`
func (p *PostgresStorage) RetrieveLatestSum(partialGitHash string) (*pb.BuildSummary, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	var sum pb.BuildSummary
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	if err = p.Connect(); err != nil {
		return &sum, errors.New("could not connect to postgres: " + err.Error())
	}
	var queuetime, starttime time.Time
	querystr := `SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE hash ilike $1 ORDER BY id DESC LIMIT 1;`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return &sum, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(partialGitHash+"%").Scan(&sum.Hash, &sum.Failed, &starttime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch, &queuetime, &sum.Status)
	if err == sql.ErrNoRows {
		ocelog.IncludeErrField(err)
		return &sum, storage_error.BuildSumNotFound(partialGitHash)
	}
	sum.BuildTime = &timestamp.Timestamp{Seconds: starttime.Unix()}
	sum.QueueTime = &timestamp.Timestamp{Seconds: queuetime.Unix()}
	return &sum, err
}

// RetrieveSumByBuildId will return a build summary based on build id
func (p *PostgresStorage) RetrieveSumByBuildId(buildId int64) (*pb.BuildSummary, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	var sum pb.BuildSummary
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	if err = p.Connect(); err != nil {
		return &sum, errors.New("could not connect to postgres: " + err.Error())
	}
	var queuetime, starttime time.Time
	querystr := `SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE id = $1`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return &sum, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(buildId).Scan(&sum.Hash, &sum.Failed, &starttime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch, &queuetime, &sum.Status)
	if err == sql.ErrNoRows {
		ocelog.IncludeErrField(err)
		return &sum, storage_error.BuildSumNotFound(string(buildId))
	}
	sum.BuildTime = &timestamp.Timestamp{Seconds: starttime.Unix()}
	sum.QueueTime = &timestamp.Timestamp{Seconds: queuetime.Unix()}
	return &sum, err
}

// RetrieveLastFewSums will return < limit> number of summaries that correlate with a repo and account.
func (p *PostgresStorage) RetrieveLastFewSums(repo string, account string, limit int32) (sums []*pb.BuildSummary, err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	if err = p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := fmt.Sprintf(`SELECT hash, failed, starttime, account, buildtime, repo, id, branch, queuetime, status FROM build_summary WHERE repo=$1 and account=$2 ORDER BY id DESC LIMIT %d`, limit)
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var queuetime, starttime time.Time
	var rows *sql.Rows
	rows, err = stmt.Query(repo, account)

	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := pb.BuildSummary{}
		if err = rows.Scan(&sum.Hash, &sum.Failed, &starttime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch, &queuetime, &sum.Status); err != nil {
			if err == sql.ErrNoRows {
				return sums, storage_error.BuildSumNotFound("account: " + account + "and repo: " + repo)
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
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	var sums []*pb.BuildSummary
	if err = p.Connect(); err != nil {
		return sums, errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := fmt.Sprintf(`select distinct on (account, repo) account, repo from build_summary where repo ilike $1;`)
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(partialRepo + "%")
	if err != nil {
		ocelog.IncludeErrField(err)
		return sums, err
	}
	defer rows.Close()
	for rows.Next() {
		sum := pb.BuildSummary{}
		if err = rows.Scan(&sum.Account, &sum.Repo); err != nil {
			if err == sql.ErrNoRows {
				return sums, storage_error.BuildSumNotFound("repository starting with" + partialRepo)
			}
			ocelog.IncludeErrField(err)
			return sums, err
		}
		sums = append(sums, &sum)
	}
	return sums, nil
}

func (p *PostgresStorage) GetTrackedRepos() (*pb.AcctRepos, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	if err = p.Connect(); err != nil {
		return nil, errors.Wrap(err, "could not connect to postgres")
	}
	var queuetime time.Time
	queryStr := `SELECT DISTINCT ON (account, repo) account, repo, queuetime
FROM build_summary
ORDER BY account, repo, queuetime DESC NULLS LAST;`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	acctRepos := &pb.AcctRepos{}
	for rows.Next() {
		acctRepo := &pb.AcctRepo{}
		if err = rows.Scan(&acctRepo.Account, &acctRepo.Repo, &queuetime); err != nil {
			if err == sql.ErrNoRows {
				return nil, storage_error.BuildSumNotFound("any account/repo")
			}
			return nil, err
		}
		acctRepo.LastQueue = convertTimeToTimestamp(queuetime)
		acctRepos.AcctRepos = append(acctRepos.AcctRepos, acctRepo)
	}
	return acctRepos, nil
}

//GetLastSuccessfulBuildHash will retrieve the last hash of a successful build on the given branch. If there are no builds
// for that branch, a storage_error.BuildSumNotFound error will be returned. If there are no successful builds,
// a storage_error.BuildSumNotFound will also be returned.
func (p *PostgresStorage) GetLastSuccessfulBuildHash(account, repo, branch string) (string, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_summary", "read")
	if err = p.Connect(); err != nil {
		return "", errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `SELECT hash 
FROM build_summary
WHERE (account,repo,branch,status) = ($1,$2,$3,$4)
ORDER BY queuetime DESC 
limit 1;`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare statment")
		return "", err
	}
	defer stmt.Close()
	row := stmt.QueryRow(account, repo, branch, pb.BuildStatus_PASSED)
	var hash string
	err = row.Scan(&hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", storage_error.BuildSumNotFound("successful build hash for branch")
		}
		return "", err
	}
	return hash, nil
}
