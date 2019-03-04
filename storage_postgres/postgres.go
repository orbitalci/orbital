package storage_postgres

import (
	"database/sql"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"

	metrics "github.com/level11consulting/ocelot/storage_metrics"
)

const TimeFormat = "2006-01-02 15:04:05"

// NewPostgresStorage take all the connection info and returns a *PostgresStorage
func NewPostgresStorage(user string, pw string, loc string, port int, dbLoc string) (*PostgresStorage, error) {

	userpass := url.UserPassword(user, pw)
	url, err := url.Parse(fmt.Sprintf("postgres://%s:%d/%s", loc, port, dbLoc))

	if err != nil {
		errorMsg := "There was an internal error parsing postgres connection info"
		return &PostgresStorage{}, errors.Wrap(err, errorMsg)
	}

	url.User = userpass

	pg := &PostgresStorage{
		url: *url,
	}

	if err := pg.Connect(); err != nil {
		return pg, err
	}
	return pg, nil
}

// url.Url and url.UserInfo
//type PostgresStorage struct {
//	user     string
//	password string
//	location string
//	port     int
//	dbLoc    string
//	db       *sql.DB
//	once     sync.Once
//}
//
type PostgresStorage struct {
	url  url.URL
	db   *sql.DB
	once sync.Once
}

func (p *PostgresStorage) Connect() error {
	p.once.Do(func() {
		var err error

		// TODO, check if password is set
		var password, _ = p.url.User.Password()

		if p.db, err = sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=disable", p.url.User.Username(), password, p.url.Host, p.url.Port())); err != nil {
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
	defer metrics.MetricizeDbErr(err)
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
	defer metrics.MetricizeDbErr(err)
	err = p.db.Close()
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


func (p *PostgresStorage) StorageType() string {
	return fmt.Sprintf("Postgres Database at %s", p.url.Host)
}
