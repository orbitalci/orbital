package postgres

import (
	"database/sql"

	"github.com/level11consulting/ocelot/models"
	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"

	metrics "github.com/level11consulting/ocelot/storage/metrics"
)

/*
  Column  |       Type        | Collation | Nullable
----------+-------------------+-----------+-----------
 build_id | integer           |           | not null
 output   | character varying |           |
 id       | integer           |           | not null
*/

//AddOut adds build output text to build_output table
func (p *PostgresStorage) AddOut(output *models.BuildOutput) error {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_output", "create")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	if err = output.Validate(); err != nil {
		ocelog.IncludeErrField(err).Error()
		return err
	}
	querystr := `INSERT INTO build_output(build_id, output) values ($1,$2)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	//"2006-01-02 15:04:05"
	if _, err = stmt.Exec(output.BuildId, output.Output); err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) RetrieveOut(buildId int64) (models.BuildOutput, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_output", "read")
	out := models.BuildOutput{}
	if err = p.Connect(); err != nil {
		return out, errors.New("could not connect to postgres: " + err.Error())
	}
	querystr := `SELECT * FROM build_output WHERE build_id=$1`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(querystr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return out, err
	}
	defer stmt.Close()
	if err = stmt.QueryRow(buildId).Scan(&out.BuildId, &out.Output, &out.OutputId); err != nil {
		ocelog.IncludeErrField(err)
		return out, err
	}
	return out, nil
}

// RetrieveLastOutByHash will return the last output text that correlates with the gitHash
func (p *PostgresStorage) RetrieveLastOutByHash(gitHash string) (models.BuildOutput, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_output", "read")
	queryStr := "select build_id, output, build_output.id from build_output " +
		"join build_summary on build_output.build_id = build_summary.id and build_summary.hash = $1 " +
		"order by build_summary.queuetime desc limit 1;"
	out := models.BuildOutput{}
	if err = p.Connect(); err != nil {
		return out, errors.New("could not connect to postgres: " + err.Error())
	}
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return out, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(gitHash).Scan(&out.BuildId, &out.Output, &out.OutputId)
	return out, err
}
