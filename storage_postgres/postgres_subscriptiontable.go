package storage_postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/level11consulting/ocelot/models/pb"
	storage_error "github.com/level11consulting/ocelot/storage/error"
	metrics "github.com/level11consulting/ocelot/storage/metrics"
	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"
)

func (p *PostgresStorage) FindSubscribeesForRepo(acctRepo string, credType pb.SubCredType) ([]*pb.ActiveSubscription, error) {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "active_subscriptions", "read")
	if err = p.Connect(); err != nil {
		return nil, errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `SELECT subscribed_to_vcs_cred_type, subscribed_to_repo, subscribing_vcs_cred_type, subscribing_repo, branch_queue_map, alias, id
FROM active_subscriptions WHERE (subscribed_to_vcs_cred_type, subscribed_to_repo)=($1,$2);`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return nil, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(credType, acctRepo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subscripts []*pb.ActiveSubscription
	for rows.Next() {
		var branchQueueBytes []byte
		sub := pb.ActiveSubscription{}
		if err = rows.Scan(&sub.SubscribedToVcsType, &sub.SubscribedToAcctRepo, &sub.SubscribingVcsType, &sub.SubscribingAcctRepo, &branchQueueBytes, &sub.Alias, &sub.Id); err != nil {
			if err == sql.ErrNoRows {
				return nil, &storage_error.ErrNotFound{Msg: "active subscription not found for " + acctRepo + credType.String()}
			}
			ocelog.IncludeErrField(err).Error("couldn't scan row")
			// todo: i guess keep trying? mayube just return
			continue
		}
		if err = json.Unmarshal(branchQueueBytes, &sub.BranchQueueMap); err != nil {
			ocelog.IncludeErrField(err).Error("unable to unmarshal branchQueueBytes to map")
			return nil, errors.Wrap(err, "unable to unmarshal branchQueueBytes to map")
		}
		subscripts = append(subscripts, &sub)
	}
	return subscripts, err

}

//InsertOrUpdateActiveSubscription will attempt to insert an active subscription in the database, and update just the BranchQueueMap & Alias if it already exists
func (p *PostgresStorage) InsertOrUpdateActiveSubscription(sub *pb.ActiveSubscription) (id int64, err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "active_subscriptions", "read")
	if err = p.Connect(); err != nil {
		return id, errors.Wrap(err, "could not connect to postgres")
	}

	queryStr := `INSERT INTO active_subscriptions(subscribed_to_vcs_cred_type, 
								  subscribed_to_repo, 
								  subscribing_vcs_cred_type, 
								  subscribing_repo, 
								  branch_queue_map,
                                  insert_time,
								  alias) 
				VALUES ($1,$2,$3,$4,$5,$6,$7) 
				ON CONFLICT(subscribed_to_repo, subscribed_to_vcs_cred_type, subscribing_repo, subscribing_vcs_cred_type) do update set branch_queue_map=EXCLUDED.branch_queue_map, alias=EXCLUDED.alias 
				RETURNING id;`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return id, err
	}
	defer stmt.Close()
	branchmap, _ := json.Marshal(sub.BranchQueueMap)
	if err = stmt.QueryRow(sub.SubscribedToVcsType, sub.SubscribedToAcctRepo, sub.SubscribingVcsType, sub.SubscribingAcctRepo, branchmap, time.Now().Format(TimeFormat), sub.Alias).Scan(&id); err != nil {
		return id, errors.Wrap(err, "unable to insert into active_subscriptions")
	}
	return
}

func (p *PostgresStorage) DeleteAllActiveSubscriptionsForRepo(acctRepo string, vcsType pb.SubCredType) (err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "active_subscriptions", "read")
	if err = p.Connect(); err != nil {
		return errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `DELETE FROM active_subscriptions 
WHERE (subscribing_repo, subscribing_vcs_cred_type)=($1,$2);`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return errors.Wrap(err, "couldn't prepare stmt")
	}
	defer stmt.Close()
	if _, err = stmt.Exec(acctRepo, vcsType); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't delete")
	}
	return
}

func (p *PostgresStorage) GetActiveSubscriptionData(subscribingAcctRepo string, subscribingBuildId int64, subscribingVcsType pb.SubCredType) (data *pb.SubscriptionUpstreamData, err error) {
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "active_subscriptions", "read")
	if err = p.Connect(); err != nil {
		return data, errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `
select 
	build_summary.branch,
	build_summary.hash,
	build_summary.account,
	build_summary.repo,
    build_summary.id,
	active_subscriptions.alias
from 
	build_summary 
inner join active_subscriptions on (CONCAT(build_summary.account, '/', build_summary.repo)) = active_subscriptions.subscribed_to_repo
where 
	build_summary.id = (select subscribed_to_build_id from subscription_data where build_id=$1)
	and active_subscriptions.subscribing_repo=$2;
`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return data, errors.Wrap(err, "couldn't prepare stmt")
	}
	defer stmt.Close()
	row := stmt.QueryRow(subscribingBuildId, subscribingAcctRepo)
	var subscriptionData = pb.SubscriptionUpstreamData{}
	if err = row.Scan(&subscriptionData.Branch, &subscriptionData.Hash, &subscriptionData.Account, &subscriptionData.Repo, &subscriptionData.BuildId, &subscriptionData.Alias); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't scan to string variables")
		return data, errors.Wrap(err, "couldn't scan to string variables")
	}
	data = &subscriptionData
	return
}

func (p *PostgresStorage) InsertSubscriptionData(upstreamBuildId, buildId, activeSubscriptionId int64) error {
	var err error
	defer metrics.MetricizeDbErr(err)
	start := metrics.StartTransaction()
	defer metrics.FinishTransaction(start, "active_subscriptions", "read")
	if err = p.Connect(); err != nil {
		return errors.Wrap(err, "could not connect to postgres")
	}
	queryStr := `INSERT INTO subscription_data (build_id, active_subscriptions_id, subscribed_to_build_id) VALUES ($1,$2,$3)`
	var stmt *sql.Stmt
	stmt, err = p.db.Prepare(queryStr)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't prepare stmt")
		return errors.Wrap(err, "couldn't prepare stmt")
	}
	defer stmt.Close()
	if _, err = stmt.Exec(buildId, activeSubscriptionId, upstreamBuildId); err != nil {
		ocelog.IncludeErrField(err).Error("unable to insert subscription_data")
		return err
	}
	return nil
}
