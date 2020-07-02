use crate::postgres::build_stage::{BuildStage, NewBuildStage};
use crate::postgres::build_summary::{BuildSummary, NewBuildSummary};
use crate::postgres::build_target::{BuildTarget, NewBuildTarget};
use crate::postgres::org::{NewOrg, Org};
use crate::postgres::repo::{NewRepo, Repo};
use crate::postgres::schema::{
    build_stage, build_summary, build_target, org, repo, secret, JobState, JobTrigger, SecretType,
};
use crate::postgres::secret::{NewSecret, Secret};
use anyhow::{anyhow, Result};
use diesel::pg::PgConnection;
use diesel::prelude::*;
use hashicorp_stack::orb_vault_path;
use log::debug;
use std::env;
//use orbital_headers::orbital_types;

pub fn establish_connection() -> PgConnection {
    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    PgConnection::establish(&database_url).expect(&format!("Error connecting to {}", database_url))
}

pub fn org_from_id(conn: &PgConnection, id: i32) -> Result<Org> {
    let org_check: Result<Org, _> = org::table
        .select(org::all_columns)
        .filter(org::id.eq(id))
        .get_result(conn);

    match org_check {
        Ok(o) => Ok(o),
        Err(_e) => Err(anyhow!("Could not retrieve org by id from DB")),
    }
}

pub fn secret_from_id(conn: &PgConnection, id: i32) -> Option<Secret> {
    let secret_check: Result<Secret, _> = secret::table
        .select(secret::all_columns)
        .filter(secret::id.eq(id))
        .get_result(conn);

    match secret_check {
        Ok(o) => Some(o),
        Err(_e) => None,
    }
}

pub fn repo_increment_build_index(conn: &PgConnection, repo: Repo) -> Result<Repo> {
    let org_name = org_from_id(&conn, repo.org_id.clone())?.name;

    let update_repo = NewRepo {
        org_id: repo.org_id,
        name: repo.name.clone(),
        uri: repo.uri,
        git_host_type: repo.git_host_type,
        secret_id: repo.secret_id,
        build_active_state: repo.build_active_state,
        notify_active_state: repo.notify_active_state,
        next_build_index: repo.next_build_index + 1,
    };

    let update_result = repo_update(conn, &org_name, &repo.name.clone(), update_repo)?;

    Ok(update_result.1)
}

pub fn org_add(conn: &PgConnection, name: &str) -> Result<Org> {
    // Only insert if there are no other orgs by this name
    let org_check: Result<Org, _> = org::table
        .select(org::all_columns)
        .filter(org::name.eq(&name))
        .order_by(org::id)
        .get_result(conn);

    match org_check {
        Err(_e) => {
            debug!("org doesn't exist. Inserting into db.");
            Ok(diesel::insert_into(org::table)
                .values(NewOrg {
                    name: name.to_string(),
                    ..Default::default()
                })
                .get_result(conn)
                .expect("Error saving new org"))
        }
        Ok(o) => {
            debug!("org found in db. Returning result.");
            Ok(o)
        }
    }
}

pub fn org_get(conn: &PgConnection, name: &str) -> Result<Org> {
    let mut org_check: Vec<Org> = org::table
        .select(org::all_columns)
        .filter(org::name.eq(&name))
        .order_by(org::id)
        .load(conn)
        .expect("Error querying for org");

    match &org_check.len() {
        0 => {
            debug!("org not found in db");
            Err(anyhow!("Org not Found"))
        }
        1 => {
            debug!("org found in db. Returning result.");
            Ok(org_check.pop().unwrap())
        }
        _ => Err(anyhow!("Found more than one org by the same name in db")),
    }
}

pub fn org_update(conn: &PgConnection, name: &str, update_org: NewOrg) -> Result<Org> {
    let org_update: Org = diesel::update(org::table)
        .filter(org::name.eq(&name))
        .set(update_org)
        .get_result(conn)
        .expect("Error updating org");

    debug!("Result after update: {:?}", &org_update);

    Ok(org_update)
}

pub fn org_remove(conn: &PgConnection, name: &str) -> Result<Org> {
    let org_delete: Org = diesel::delete(org::table)
        .filter(org::name.eq(&name))
        .get_result(conn)
        .expect("Error deleting org");

    Ok(org_delete)
}

pub fn org_list(conn: &PgConnection) -> Result<Vec<Org>> {
    let query: Vec<Org> = org::table
        .select(org::all_columns)
        .order_by(org::id)
        .load(conn)
        .expect("Error getting all order records");

    Ok(query)
}

pub fn secret_add(
    conn: &PgConnection,
    org: &str,
    name: &str,
    secret_type: SecretType,
) -> Result<(Secret, Org)> {
    let query_result: Result<(Secret, Org), _> = secret::table
        .inner_join(org::table)
        .select((secret::all_columns, org::all_columns))
        .filter(secret::name.eq(&name))
        .filter(org::name.eq(&org))
        .get_result(conn);

    match query_result {
        Err(_e) => {
            debug!("secret doesn't exist. Inserting into db.");

            let org_db = org_get(conn, org).expect("Unable to find org");

            let secret_db = diesel::insert_into(secret::table)
                .values(NewSecret {
                    name: name.to_string(),
                    org_id: org_db.id,
                    secret_type: secret_type,
                    vault_path: orb_vault_path(
                        &org_db.name,
                        name,
                        format!("{:?}", &secret_type).as_str(),
                    ),
                    ..Default::default()
                })
                .get_result(conn)
                .expect("Error saving new secret");

            Ok((secret_db, org_db))
        }
        Ok((secret_db, org_db)) => {
            debug!("secret found in db. Returning result.");
            Ok((secret_db, org_db))
        }
    }
}

pub fn secret_get(
    conn: &PgConnection,
    org: &str,
    name: &str,
    _secret_type: SecretType,
) -> Result<(Secret, Org)> {
    let query_result: (Secret, Org) = secret::table
        .inner_join(org::table)
        .select((secret::all_columns, org::all_columns))
        .filter(secret::name.eq(&name))
        .filter(org::name.eq(&org))
        .first(conn)
        .expect("Error querying for secret");

    debug!("Secret get result: \n {:?}", &query_result);

    Ok(query_result)
}

pub fn secret_update(
    conn: &PgConnection,
    org: &str,
    name: &str,
    update_secret: NewSecret,
) -> Result<Secret> {
    let org_db = org_get(conn, org).expect("Unable to find org");

    let secret_update: Secret = diesel::update(secret::table)
        .filter(secret::org_id.eq(&org_db.id))
        .filter(secret::name.eq(&name))
        .set(update_secret)
        .get_result(conn)
        .expect("Error updating secret");

    debug!("Result after update: {:?}", &secret_update);

    Ok(secret_update)
}

pub fn secret_remove(
    conn: &PgConnection,
    org: &str,
    name: &str,
    _secret_type: SecretType,
) -> Result<Secret> {
    let org_db = org_get(conn, org).expect("Unable to find org");

    let secret_delete: Secret = diesel::delete(secret::table)
        .filter(secret::org_id.eq(&org_db.id))
        .filter(secret::name.eq(&name))
        .get_result(conn)
        .expect("Error deleting secret");

    Ok(secret_delete)
}

pub fn secret_list(
    conn: &PgConnection,
    org: &str,
    filter: Option<SecretType>,
) -> Result<Vec<(Secret, Org)>> {
    let query_result: Vec<(Secret, Org)> = match filter {
        None => secret::table
            .inner_join(org::table)
            .select((secret::all_columns, org::all_columns))
            .filter(org::name.eq(&org))
            .load(conn)
            .expect("Error getting all secret records"),
        Some(_f) => secret::table
            .inner_join(org::table)
            .select((secret::all_columns, org::all_columns))
            .filter(org::name.eq(&org))
            //.filter(secret::secret_type.eq(SecretType::from(f))) // Not working yet.
            .load(conn)
            .expect("Error getting all secret records"),
    };

    debug!("Secret list result: \n {:?}", &query_result);

    Ok(query_result)
}

pub fn repo_add(
    conn: &PgConnection,
    org: &str,
    name: &str,
    uri: &str,
    secret: Option<Secret>,
) -> Result<(Org, Repo, Option<Secret>)> {
    let repo_check = repo_get(conn, org, name);

    match repo_check {
        Err(_e) => {
            debug!("repo doesn't exist. Inserting into db.");

            let secret_id = match secret {
                Some(s) => Some(s.clone().id),
                None => None,
            };

            let org_db = org_get(conn, org)?;

            let result: Repo = diesel::insert_into(repo::table)
                .values(NewRepo {
                    name: name.into(),
                    org_id: org_db.id,
                    uri: uri.into(),
                    secret_id: secret_id,
                    ..Default::default()
                })
                .get_result(conn)
                .expect("Error saving new repo");

            debug!("DB insert result: {:?}", &result);

            // Run query again. This time it should pass
            repo_get(conn, org, name)
        }
        Ok((o, r, s)) => {
            debug!("repo found in db. Returning result.");
            Ok((o, r, s))
        }
    }
}

pub fn repo_get(conn: &PgConnection, org: &str, name: &str) -> Result<(Org, Repo, Option<Secret>)> {
    debug!("Repo get: Org: {}, Name: {}", org, name);

    let query: Result<(Org, Repo), _> = repo::table
        .inner_join(org::table)
        .select((org::all_columns, repo::all_columns))
        .filter(repo::name.eq(&name))
        //.filter(secret::id.eq(&secret_id))
        .get_result(conn);

    match query {
        Ok((o, r)) => {
            // If we're using a secret, we should also return it to the requester
            let secret = match &r.secret_id {
                None => None,
                Some(id) => secret_from_id(conn, *id),
            };

            Ok((o, r, secret))
        }
        Err(_e) => Err(anyhow!("{} not found in {} org", name, org)),
    }
}

// You should update your secret with secret_update()
pub fn repo_update(
    conn: &PgConnection,
    org: &str,
    name: &str,
    update_repo: NewRepo,
) -> Result<(Org, Repo, Option<Secret>)> {
    let (org_db, _repo_db, secret_db) = repo_get(conn, org, name)?;

    let repo_update: Repo = diesel::update(repo::table)
        .filter(repo::org_id.eq(&org_db.id))
        .filter(repo::name.eq(&name))
        .set(update_repo)
        .get_result(conn)
        .expect("Error updating repo");

    debug!("Result after update: {:?}", &repo_update);

    Ok((org_db, repo_update, secret_db))
}

pub fn repo_remove(
    conn: &PgConnection,
    org: &str,
    name: &str,
) -> Result<(Org, Repo, Option<Secret>)> {
    let (org_db, repo_db, secret_db) = repo_get(conn, org, name)?;

    let _repo_delete: Repo = diesel::delete(repo::table)
        .filter(repo::org_id.eq(&org_db.id))
        .filter(repo::name.eq(&name))
        .get_result(conn)
        .expect("Error deleting repo");

    Ok((org_db, repo_db, secret_db))
}

pub fn repo_list(conn: &PgConnection, org: &str) -> Result<Vec<(Org, Repo, Option<Secret>)>> {
    let query: Vec<(Org, Repo)> = repo::table
        .inner_join(org::table)
        .select((org::all_columns, repo::all_columns))
        .filter(org::name.eq(org))
        .load(conn)
        .expect("Error selecting all repo");

    let map_result: Vec<(Org, Repo, Option<Secret>)> = query
        .into_iter()
        .map(|(o, r)| match r.clone().secret_id {
            None => (o, r, None),
            Some(id) => (o, r, secret_from_id(conn, id)),
        })
        .collect();

    Ok(map_result)
}

pub fn build_target_add(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    user_envs: Option<String>,
    job_trigger: JobTrigger,
) -> Result<(Org, Repo, BuildTarget)> {
    let (org_db, repo_db, _) = repo_get(conn, org, repo)?;

    let build_target = NewBuildTarget {
        repo_id: repo_db.id.clone(),
        git_hash: hash.to_string(),
        branch: branch.to_string(),
        user_envs: user_envs,
        build_index: repo_db.next_build_index,
        trigger: job_trigger,
        ..Default::default()
    };

    debug!("Build spec to insert: {:?}", &build_target);

    let result: BuildTarget = diesel::insert_into(build_target::table)
        .values(build_target)
        .get_result(conn)
        .expect("Error saving new build_target");

    // Increment repo next_build_target by 1
    let updated_repo = repo_increment_build_index(conn, repo_db)?;

    Ok((org_db, updated_repo, result))
}

// This should probably return a Vec
// Consider taking a Repo as input
pub fn build_target_get(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
) -> Result<(Org, Repo, Option<BuildTarget>)> {
    debug!(
        "Build target get request: org {:?} repo: {:?} hash: {:?} branch: {:?} build_index: {:?}",
        &org, &repo, &hash, &branch, &build_index,
    );

    let (org_db, repo_db, _secret_db) = repo_get(conn, org, repo)?;

    let result: Result<(Repo, BuildTarget), _> = build_target::table
        .inner_join(repo::table)
        .select((repo::all_columns, build_target::all_columns))
        .filter(build_target::repo_id.eq(repo_db.id))
        .filter(build_target::branch.eq(branch))
        .filter(build_target::build_index.eq(build_index))
        .get_result(conn);

    match result {
        Ok((repo, build_target)) => {
            debug!("BuildTarget found: {:?}", &build_target);
            Ok((org_db, repo, Some(build_target)))
        }
        Err(_e) => Ok((org_db, repo_db, None)),
    }
}

pub fn build_target_update(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
    update_build_target: NewBuildTarget,
) -> Result<(Org, Repo, BuildTarget)> {
    let (org_db, repo_db, build_target_db_opt) =
        build_target_get(conn, org, repo, hash, branch, build_index)?;

    let build_target_db = build_target_db_opt.expect("No build target found");

    let result: BuildTarget = diesel::update(build_target::table)
        .filter(build_target::id.eq(build_target_db.id))
        .set(update_build_target)
        .get_result(conn)
        .expect("Error updating build target");

    Ok((org_db, repo_db, result))
}

pub fn build_target_remove() {
    unimplemented!();
}

pub fn build_target_list(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    limit: i32,
) -> Result<Vec<(Org, Repo, BuildTarget)>> {
    debug!(
        "Build target list request: org {:?} repo: {:?} limit: {:?}",
        &org, &repo, &limit
    );

    let (org_db, _repo_db, _secret_db) = repo_get(conn, org, repo)?;

    let result: Vec<(Repo, BuildTarget)> = build_target::table
        .inner_join(repo::table)
        .select((repo::all_columns, build_target::all_columns))
        .limit(limit.into())
        .load(conn)
        .expect("Error saving new build_target");

    let map_result: Vec<(Org, Repo, BuildTarget)> = result
        .into_iter()
        .map(|(r, b)| (org_db.clone(), r, b))
        .collect();

    Ok(map_result)
}

pub fn build_summary_add(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
    build_summary: NewBuildSummary,
) -> Result<(Repo, BuildTarget, BuildSummary)> {
    debug!(
        "Build summary add request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?}",
        &org, &repo, &hash, &branch, &build_index,
    );

    let (_org_db, repo_db, build_target_db_opt) =
        build_target_get(conn, org, repo, hash, branch, build_index)?;

    let build_target_db = build_target_db_opt.expect("Build target not found");

    debug!("Build summary to insert: {:?}", &build_summary);

    let result: BuildSummary = diesel::insert_into(build_summary::table)
        .values(build_summary)
        .get_result(conn)
        .expect("Error saving new build_summary");

    Ok((repo_db, build_target_db, result))
}

pub fn build_summary_get(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
) -> Result<(Repo, BuildTarget, Option<BuildSummary>)> {
    debug!(
        "Build summary get request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?}",
        &org, &repo, &hash, &branch, &build_index,
    );

    let (_org_db, repo_db, build_target_db_opt) =
        build_target_get(conn, org, repo, hash, branch, build_index)?;

    let build_target_db = build_target_db_opt.expect("No build target found");

    let result: Result<(BuildTarget, BuildSummary), _> = build_summary::table
        .inner_join(build_target::table)
        .select((build_target::all_columns, build_summary::all_columns))
        .filter(build_summary::build_target_id.eq(build_target_db.id))
        .get_result(conn);

    match result {
        Ok((build_target, build_summary)) => {
            debug!("Build summary was found: {:?}", &build_summary);
            Ok((repo_db, build_target, Some(build_summary)))
        }
        Err(_e) => {
            debug!("Build summary NOT found");
            Ok((repo_db, build_target_db, None))
        }
    }
}

pub fn build_summary_update(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
    update_summary: NewBuildSummary,
) -> Result<(Repo, BuildTarget, BuildSummary)> {
    debug!(
        "Build summary update request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} update_summary: {:?}",
        &org, &repo, &hash, &branch, &build_index, &update_summary,
    );

    let (org_db, build_target_db, build_summary_db_opt) =
        build_summary_get(conn, org, repo, hash, branch, build_index)?;

    let build_summary_db = build_summary_db_opt.expect("No build summary found");

    let result: BuildSummary = diesel::update(build_summary::table)
        .filter(build_summary::id.eq(build_summary_db.id))
        .set(update_summary)
        .get_result(conn)
        .expect("Error updating build summary");

    Ok((org_db, build_target_db, result))
}

pub fn build_summary_remove() {
    unimplemented!();
}

// TODO: `repo` should be changed to Option<&str> for granularity between all or one repo
pub fn build_summary_list(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    limit: i32,
) -> Result<Vec<(Repo, BuildTarget, BuildSummary)>> {
    debug!(
        "Build summary list request: org {:?} repo: {:?} limit: {:?}",
        &org, &repo, &limit
    );

    let (_org_db, repo_db, _secret_db) = repo_get(conn, org, repo)?;

    let result: Vec<(BuildTarget, BuildSummary)> = build_summary::table
        .inner_join(build_target::table)
        .select((build_target::all_columns, build_summary::all_columns))
        .filter(build_target::repo_id.eq(repo_db.id))
        .order(build_summary::id.desc())
        .limit(limit.into())
        .load(conn)
        .expect("Error listing build summaries");

    let map_result: Vec<(Repo, BuildTarget, BuildSummary)> = result
        .into_iter()
        .map(|(build_target, build_summary)| (repo_db.clone(), build_target, build_summary))
        .collect();

    Ok(map_result)
}

pub fn build_stage_add(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
    build_summary_id: i32,
    build_stage: NewBuildStage,
) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
    debug!(
        "Build stage add request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} build_summary_id {:?}",
        &org, &repo, &hash, &branch, &build_index, &build_summary_id,
    );

    let (_org_db, build_target_db, build_summary_db_opt) =
        build_summary_get(conn, org, repo, hash, branch, build_index)?;

    let build_summary_db = build_summary_db_opt.expect("No build summary found");

    debug!("Build stage to insert: {:?}", &build_stage);

    let result: BuildStage = diesel::insert_into(build_stage::table)
        .values(build_stage)
        .get_result(conn)
        .expect("Error saving new build_stage");

    Ok((build_target_db, build_summary_db, result))
}

pub fn build_stage_get(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
    build_summary_id: i32,
    build_stage_id: i32,
) -> Result<(BuildTarget, BuildSummary, Option<BuildStage>)> {
    debug!(
        "Build stage get request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} build_summary_id {:?}",
        &org, &repo, &hash, &branch, &build_index, &build_summary_id,
    );

    let (_repo_db, build_target_db, build_summary_db_opt) =
        build_summary_get(conn, org, repo, hash, branch, build_index)?;

    let build_summary_db = build_summary_db_opt.expect("No build target found");

    let result: Result<(BuildSummary, BuildStage), _> = build_stage::table
        .inner_join(build_summary::table)
        .select((build_summary::all_columns, build_stage::all_columns))
        .filter(build_summary::build_target_id.eq(build_target_db.id))
        .filter(build_stage::id.eq(build_stage_id))
        .get_result(conn);

    match result {
        Ok((build_summary, build_stage)) => {
            debug!("Build stage was found: {:?}", &build_summary);
            Ok((build_target_db, build_summary, Some(build_stage)))
        }
        Err(_e) => {
            debug!("Build stage NOT found");
            Ok((build_target_db, build_summary_db, None))
        }
    }
}

pub fn build_stage_update(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
    build_summary_id: i32,
    build_stage_id: i32,
    update_stage: NewBuildStage,
) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
    debug!(
        "Build stage update request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} build_summary_id {:?} build_stage_id {:?} update_stage {:?}",
        &org, &repo, &hash, &branch, &build_index, &build_summary_id, &build_stage_id, &update_stage,
    );

    let (build_target_db, build_summary_db, build_stage_db_opt) = build_stage_get(
        conn,
        org,
        repo,
        hash,
        branch,
        build_index,
        build_summary_id,
        build_stage_id,
    )?;

    let _build_stage_db = build_stage_db_opt.expect("No build stage found");

    let result: BuildStage = diesel::update(build_stage::table)
        .filter(build_stage::id.eq(build_stage_id))
        .set(update_stage)
        .get_result(conn)
        .expect("Error updating build stage");

    Ok((build_target_db, build_summary_db, result))
}

pub fn build_stage_remove() {
    unimplemented!();
}

pub fn build_stage_list(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
    limit: i32,
) -> Result<Vec<(BuildTarget, BuildSummary, BuildStage)>> {
    debug!(
        "Build stage list request: org {:?} repo: {:?} hash {:?} branch {:?} build_index {:?} limit: {:?}",
        &org, &repo, &hash, &branch, &build_index, &limit
    );

    let (_repo_db, build_target_db, build_summary_db_opt) =
        build_summary_get(conn, org, repo, hash, branch, build_index)?;

    let build_summary_db = build_summary_db_opt.expect("No build summary found");

    let result: Vec<(BuildSummary, BuildStage)> = build_stage::table
        .inner_join(build_summary::table)
        .select((build_summary::all_columns, build_stage::all_columns))
        .filter(build_summary::build_target_id.eq(build_target_db.id))
        .filter(build_stage::build_summary_id.eq(build_summary_db.id))
        .order(build_stage::id.asc())
        .limit(limit.into())
        .load(conn)
        .expect("Error listing build stages");

    debug!(
        "Found {} stages for build id {}",
        &result.len(),
        build_index
    );

    let map_result: Vec<(BuildTarget, BuildSummary, BuildStage)> = result
        .into_iter()
        .map(|(build_summary, build_stage)| (build_target_db.clone(), build_summary, build_stage))
        .collect();

    Ok(map_result)
}

pub fn build_logs_get(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: Option<i32>,
) -> Result<Vec<(BuildTarget, BuildSummary, BuildStage)>> {
    let (_org_db, repo_db, _secret_db) = repo_get(conn, org, repo)?;

    match build_index {
        Some(n) => build_stage_list(conn, org, repo, hash, branch, n, 255),
        None => build_stage_list(
            conn,
            org,
            repo,
            hash,
            branch,
            repo_db.next_build_index - 1,
            255,
        ),
    }
}

pub fn is_build_cancelled(
    conn: &PgConnection,
    org: &str,
    repo: &str,
    hash: &str,
    branch: &str,
    build_index: i32,
) -> Result<bool> {
    match build_summary_get(conn, org, repo, hash, branch, build_index) {
        Ok((_, _, Some(summary))) => match summary.build_state {
            JobState::Cancelled => Ok(true),
            _ => Ok(false),
        },
        Ok((_, _, None)) => {
            // Build hasn't been queued yet
            Ok(false)
        }
        Err(_) => Err(anyhow!("Could not retrieve build summary from DB")),
    }
}
