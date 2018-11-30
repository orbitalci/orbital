package storage

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/lib/pq"
	ocelog "github.com/shankj3/go-til/log"
<<<<<<< HEAD
	"github.com/level11consulting/ocelot/common"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
=======
>>>>>>> trying to revert everything
)

const TimeFormat = "2006-01-02 15:04:05"

func NewPostgresStorage(user string, pw string, loc string, port int, dbLoc string) *PostgresStorage {
	pg := &PostgresStorage{
		user:     user,
		password: pw,
		location: loc,
		port:     port,
		dbLoc:    dbLoc,
	}
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
		p.db.SetMaxIdleConns(5)
	})
	return nil
}

// todo: need to write a test for this
func (p *PostgresStorage) Healthy() bool {
	var err error
	defer metricizeDbErr(err)
	err = p.Connect()
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
	var err error
	defer metricizeDbErr(err)
	err = p.db.Close()
	if err != nil {
		ocelog.IncludeErrField(err).Error("error closing postgres db")
	}
}

func convertTimeToTimestamp(tyme time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{Seconds: tyme.Unix()}
}

func (p *PostgresStorage) StorageType() string {
	return fmt.Sprintf("Postgres Database at %s", p.location)
}
