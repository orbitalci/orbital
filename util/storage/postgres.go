package storage


import (
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

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
		return err
	}
	p.db = db
	return nil
}

func (p *PostgresStorage) Disconnect() {
	p.db.Close()
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
	//"2006-01-02 15:04:05"
	if err := p.db.QueryRow(`INSERT INTO build_summary(hash, starttime, account, repo, branch, failed, buildtime) values ($1,$2,$3,$4,$5,true,-99.999) RETURNING id`,
		hash, starttime.Format("2006-01-02 15:04:05"), account, repo, branch).Scan(&id); err != nil {
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
			return sums, err
		}
		//fmt.Println(hi)
		sums = append(sums, sum)
	}
	return sums, nil
}

//RetrieveHashStartsWith will return a list of all hashes starting with the partial string in db
func (p *PostgresStorage) RetrieveHashStartsWith(partialGitHash string) ([]models.BuildSummary, error) {
	var hashes []models.BuildSummary
	if err := p.Connect(); err != nil {
		return hashes, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	rows, err := p.db.Query(`select distinct (hash), account, repo from build_summary where hash ilike $1`, partialGitHash + "%")
	if err != nil {
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

// RetrieveLatestSum will return the latest entry of build_summary where hash=gitHash
func (p *PostgresStorage) RetrieveLatestSum(gitHash string) (models.BuildSummary, error) {
	var sum models.BuildSummary
	if err := p.Connect(); err != nil {
		return sum, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()
	querystr := `SELECT * FROM build_summary WHERE hash = $1 ORDER BY starttime DESC LIMIT 1`
	row := p.db.QueryRow(querystr, gitHash)
	err := row.Scan(&sum.Hash, &sum.Failed, &sum.BuildTime, &sum.Account, &sum.BuildDuration, &sum.Repo, &sum.BuildId, &sum.Branch)
	if err == sql.ErrNoRows {
		return sum, BuildSumNotFound(gitHash)
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
	//fmt.Println(queryStr, string(buildId))
	if err := p.db.QueryRow(queryStr, buildId).Scan(&out.BuildId, &out.Output, &out.OutputId); err != nil {
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

//
func (p *PostgresStorage) AddFail(reason *models.BuildFailureReason) error {
	if err := p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()

	if err := reason.Validate(); err != nil {
		return err
	}
	queryStr := `INSERT INTO build_failure_reason(build_id, reasons) values($1, $2)`
	jsonStr, err := json.Marshal(reason.FailureReasons)
	if err != nil {
		return err
	}
	if _, err := p.db.Query(queryStr, reason.BuildId, string(jsonStr)); err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) RetrieveFail(buildId int64) (models.BuildFailureReason, error) {
	out := models.BuildFailureReason{}
	if err := p.Connect(); err != nil {
		return out, errors.New("could not connect to postgres: " + err.Error())
	}
	defer p.Disconnect()

	queryStr := `SELECT * FROM build_failure_reason WHERE build_id=$1`
	if err := p.db.QueryRow(queryStr, buildId).Scan(&out.BuildId, &out.FailureReasons, &out.FailureReasonId); err != nil {
		return out, err
	}
	return out, nil

}

func (p *PostgresStorage) StorageType() string {
	return fmt.Sprintf("Postgres Database at %s", p.location)
}