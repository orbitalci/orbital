package storage_postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/level11consulting/ocelot/build/helpers/stringbuilder"
	"github.com/level11consulting/ocelot/models/pb"
	metrics "github.com/level11consulting/ocelot/storage/metrics"
	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"
)

func (p *PostgresStorage) InsertPoll(account string, repo string, cronString string, branches string, credsId int64) (err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "poll_table", "create")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `INSERT INTO polling_repos(account, repo, cron_string, branches, last_cron_time, credentials_id) values ($1, $2, $3, $4, $5, $6)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(account, repo, cronString, branches, time.Now(), credsId); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).WithField("cronString", cronString).Error("could not insert poll entry into database")
		return
	}
	return
}

func (p *PostgresStorage) UpdatePoll(account string, repo string, cronString string, branches string) (err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "poll_table", "update")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `UPDATE polling_repos SET (cron_string, branches) = ($1,$2) WHERE (account,repo) = ($3,$4);`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
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
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "poll_table", "update")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	//starttime.Format(TimeFormat)
	var hashes []byte
	hashes, err = json.Marshal(lasthashes)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't marshal hash map to byte??")
		return err
	}
	queryStr := `UPDATE polling_repos SET (last_cron_time, last_hashes)=($1,$2) WHERE (account,repo) = ($3,$4);`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(time.Now().Format(TimeFormat), hashes, account, repo); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).Error("could not update last_cron_time in database")
	}
	return
}

func (p *PostgresStorage) GetLastData(accountRepo string) (timestamp time.Time, hashes map[string]string, err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "poll_table", "read")
	if err = p.Connect(); err != nil {

		return time.Now(), nil, errors.New("could not connect to postgres: " + err.Error())
	}
	account, repo, err := stringbuilder.GetAcctRepo(accountRepo)
	if err != nil {
		return time.Now(), nil, err
	}
	queryStr := `SELECT last_cron_time, last_hashes FROM polling_repos WHERE (account,repo) = ($1,$2);`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
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
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "poll_table", "read")
	if err = p.Connect(); err != nil {
		return false, errors.New("could not connect to postgres: " + err.Error())
	}
	var count int64
	queryStr := `select count(*) from polling_repos where (account,repo) = ($1,$2);`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
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
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "poll_table", "delete")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `delete from polling_repos where (account, repo) =($1,$2)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(account, repo); err != nil {
		ocelog.IncludeErrField(err).WithField("account", account).WithField("repo", repo).Error("could not delete poll entry from database")
		return err
	}
	return nil
}

func (p *PostgresStorage) GetAllPolls() ([]*pb.PollRequest, error) {

	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "poll_table", "read")
	var polls []*pb.PollRequest
	if err = p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `SELECT polling_repos.account, polling_repos.repo, polling_repos.cron_string, polling_repos.last_cron_time, polling_repos.branches, credentials.cred_sub_type 
FROM polling_repos LEFT JOIN credentials
	ON credentials_id = id;`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		pr := &pb.PollRequest{}
		var tyme time.Time
		if err = rows.Scan(&pr.Account, &pr.Repo, &pr.Cron, &tyme, &pr.Branches, &pr.Type); err != nil {
			return nil, err
		}
		pr.LastCronTime = convertTimeToTimestamp(tyme)
		polls = append(polls, pr)
	}
	return polls, rows.Err()
}
