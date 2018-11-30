package storage

import (
	"database/sql"

	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/models/pb"
)


//InsertCred will insert an ocyCredder object into the credentials table after calling its ValidateForInsert method.
// if the OcyCredder fails validation, it will return a *models.ValidationErr
func (p *PostgresStorage) InsertCred(credder pb.OcyCredder, overwriteOk bool) error {
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "create")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	if invalid := credder.ValidateForInsert(); invalid != nil {
		return invalid
	}
	//possibleCred, err := p.RetrieveCred(credder.GetSubType(), identifier, accountName string)
	var moreFields []byte
	moreFields, err = credder.CreateAdditionalFields()
	if err != nil {
		return errors.New("could not create additional_fields column, error: " + err.Error())
	}
	queryStr := `INSERT INTO credentials(account, identifier, cred_type, cred_sub_type, additional_fields) values ($1,$2,$3,$4,$5)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType().Parent(), credder.GetSubType(), moreFields)
	return err
}

func (p *PostgresStorage) UpdateCred(credder pb.OcyCredder) error {
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "update")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	if invalid := credder.ValidateForInsert(); invalid != nil {
		return invalid
	}
	var moreFields []byte
	moreFields, err = credder.CreateAdditionalFields()
	if err != nil {
		return errors.New("could not create additional_fields column, error: " + err.Error())
	}
	queryStr := `UPDATE credentials SET additional_fields=$1 WHERE (account,identifier,cred_sub_type)=($2,$3,$4)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(moreFields, credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType())
	return err
}

func (p *PostgresStorage) DeleteCred(credder pb.OcyCredder) error {
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "delete")
	if err = p.Connect(); err != nil {
		return errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `DELETE from credentials where (account,identifier,cred_sub_type)=($1,$2,$3)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare statement")
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType())
	return err

}

func (p *PostgresStorage) CredExists(credder pb.OcyCredder) (bool, error) {
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err = p.Connect(); err != nil {
		return false, errors.New("could not connect to postgres: " + err.Error())
	}
	var count int64
	queryStr := `select count(*) from credentials where (account,identifier,cred_sub_type) = ($1,$2,$3);`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
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

func scanCredRowToCredder(rows *sql.Rows) (pb.OcyCredder, error) {
	var credType, subCredType int32
	var addtlFields []byte
	var account, identifier string
	var id int64
	err := rows.Scan(&account, &identifier, &credType, &subCredType, &addtlFields, &id)
	if err != nil {
		return nil, err
	}
	ocyCredType := pb.CredType(credType)
	cred := ocyCredType.SpawnCredStruct(account, identifier, pb.SubCredType(subCredType), id)
	if err = cred.UnmarshalAdditionalFields(addtlFields); err != nil {
		return nil, err
	}
	return cred, nil
}

func (p *PostgresStorage) RetrieveAllCreds() ([]pb.OcyCredder, error) {
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err = p.Connect(); err != nil {
		return nil, errors.New("could not connect to postgres: " + err.Error())
	}
	queryStr := `SELECT account, identifier, cred_type, cred_sub_type, additional_fields, id from credentials order by cred_type`
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
	var creds []pb.OcyCredder
	for rows.Next() {
		var cred pb.OcyCredder
		cred, err = scanCredRowToCredder(rows)
		if err != nil {
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
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err = p.Connect(); err != nil {
		return nil, errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `SELECT account, identifier, cred_type, cred_sub_type, additional_fields, id FROM credentials WHERE cred_type=$1`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(credType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []pb.OcyCredder
	for rows.Next() {
		var cred pb.OcyCredder
		cred, err = scanCredRowToCredder(rows)
		if err != nil {
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
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err = p.Connect(); err != nil {
		return nil, errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `SELECT additional_fields, id FROM credentials WHERE (cred_sub_type,identifier,account)=($1,$2,$3)`
	ocelog.Log().Debugf("%d %s %s", subCredType, identifier, accountName)
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var addtlFields []byte
	var id int64
	if err = stmt.QueryRow(subCredType, identifier, accountName).Scan(&addtlFields, &id); err != nil {
		if err == sql.ErrNoRows {
			return nil, CredNotFound(accountName, identifier)
		}
		return nil, err
	}
	credder := subCredType.Parent().SpawnCredStruct(accountName, identifier, subCredType, id)
	if credder == nil {
		// do we even need this check? wouldn't strict typing never allow this condition?
		return nil, errors.New("credder is nil")
	}
	err = credder.UnmarshalAdditionalFields(addtlFields)
	return credder, err
}

func (p *PostgresStorage) RetrieveCredBySubTypeAndAcct(scredType pb.SubCredType, acctName string) ([]pb.OcyCredder, error) {
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err = p.Connect(); err != nil {
		return nil, errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `SELECT additional_fields, identifier, id FROM credentials WHERE (cred_sub_type,account)=($1,$2)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(scredType, acctName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []pb.OcyCredder
	for rows.Next() {
		var addtlFields []byte
		var identifier string
		var id int64
		err = rows.Scan(&addtlFields, &identifier, &id)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, CredNotFound(acctName, scredType.String())
			}
			return nil, err
		}
		credder := scredType.Parent().SpawnCredStruct(acctName, identifier, scredType, id)
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

func (p *PostgresStorage) GetVCSTypeFromAccount(account string) (pb.SubCredType, error){
	var bad pb.SubCredType = pb.SubCredType_NIL_SCT

	//  I'm not sure if this form is compilable
	// const pb.SubCredType bad = pb.SubCredType_NIL_SCT
	var err error
	defer metricizeDbErr(err)
	start := startTransaction()
	defer finishTransaction(start, "credentials", "read")
	if err = p.Connect(); err != nil {
		return bad, errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `SELECT DISTINCT cred_sub_type FROM credentials WHERE 
(cred_type,account)=($1,$2)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return bad, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(pb.CredType_VCS, account)
	if err != nil {
		return bad, err
	}
	defer rows.Close()
	var scts []pb.SubCredType
	for rows.Next() {
		sct := pb.SubCredType_NIL_SCT
		err = rows.Scan(&sct)
		if err != nil {
			if err == sql.ErrNoRows {
				return bad, CredNotFound(account, "any")
			}
			return bad, err
		}
		scts = append(scts, sct)
	}
	if len(scts) != 1 {
		return bad, MultipleVCSTypes(account, scts)
	}
	return scts[0], nil
}
