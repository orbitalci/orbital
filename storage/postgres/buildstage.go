package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/level11consulting/orbitalci/models"
	storage_error "github.com/level11consulting/orbitalci/storage/error"
	metrics "github.com/level11consulting/orbitalci/storage/metrics"
	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"
)

// AddStageDetail will store the stage data along with a starttime and duration to db
//  The fields required on stageResult to insert into stage detail table are:
// 		stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime, stageResult.StageDuration, stageResult.Status, stageResult.Messages
func (p *PostgresStorage) AddStageDetail(stageResult *models.StageResult) error {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_stage_details", "create")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	if err = stageResult.Validate(); err != nil {
		ocelog.IncludeErrField(err)
		return err
	}
	queryStr := `INSERT INTO build_stage_details(build_id, stage, error, starttime, runtime, status, messages) values ($1, $2, $3, $4, $5, $6, $7)`
	var jsonStr []byte
	jsonStr, err = json.Marshal(stageResult.Messages)
	if err != nil {
		ocelog.IncludeErrField(err)
		return err
	}
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
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
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "build_stage_details", "read")
	var stages []models.StageResult
	queryStr := "select id, build_id, error, starttime, runtime, status, messages, stage from build_stage_details where build_id = $1 order by build_id asc;"
	if err = p.Connect(); err != nil {
		return stages, errors.New("could not connect to postgres: " + err.Error())
	}
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(buildId)
	defer rows.Close()
	for rows.Next() {
		stage := models.StageResult{}
		var errString sql.NullString        //using sql's NullString because calling .Scan
		var messages models.JsonStringArray //have to use custom class because messages are stored in json format

		if err = rows.Scan(&stage.StageResultId, &stage.BuildId, &errString, &stage.StartTime, &stage.StageDuration, &stage.Status, &messages, &stage.Stage); err != nil {
			if err == sql.ErrNoRows {
				return stages, storage_error.StagesNotFound(fmt.Sprintf("build id: %v", buildId))
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
