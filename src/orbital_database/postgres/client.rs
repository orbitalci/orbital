use crate::orbital_database::postgres::build_stage::{BuildStage, NewBuildStage};
use crate::orbital_database::postgres::build_summary::{BuildSummary, NewBuildSummary};
use crate::orbital_database::postgres::build_target::{BuildTarget, NewBuildTarget};
use crate::orbital_database::postgres::org::{NewOrg, Org};
use crate::orbital_database::postgres::repo::{NewRepo, Repo};
use crate::orbital_database::postgres::schema::{
    build_stage, build_summary, build_target, org, repo, secret, JobState, JobTrigger, SecretType,
};
use crate::orbital_database::postgres::secret::{NewSecret, Secret};
use crate::orbital_utils::hashicorp_stack::orb_vault_path;
use color_eyre::eyre::{eyre, Result};
use diesel::pg::PgConnection;
use diesel::prelude::*;
use log::debug;
use std::env;

pub struct OrbitalDBClient {
    conn: PgConnection,
    org: Option<String>,
    repo: Option<String>,
    branch: Option<String>,
    hash: Option<String>,
}

impl OrbitalDBClient {
    pub fn new() -> Self {
        let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
        Self {
            conn: PgConnection::establish(&database_url)
                .unwrap_or_else(|_| panic!("Error connecting to {}", database_url)),
            org: None,
            repo: None,
            branch: None,
            hash: None,
        }
    }

    pub fn get_conn(&self) -> PgConnection {
        let conn = OrbitalDBClient::new();
        conn.conn
    }

    pub fn set_org(mut self, org_name: Option<String>) -> Self {
        self.org = org_name;
        self
    }

    pub fn set_repo(mut self, repo_name: Option<String>) -> Self {
        self.repo = repo_name;
        self
    }

    pub fn set_branch(mut self, branch_name: Option<String>) -> Self {
        self.branch = branch_name;
        self
    }

    pub fn set_hash(mut self, hash: Option<String>) -> Self {
        self.hash = hash;
        self
    }
    //pub fn establish_connection() -> PgConnection {
    //    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    //    PgConnection::establish(&database_url)
    //        .unwrap_or_else(|_| panic!("Error connecting to {}", database_url))
    //}

    pub fn org_from_id(&self, id: i32) -> Result<Org> {
        let org_check: Result<Org, _> = org::table
            .select(org::all_columns)
            .filter(org::id.eq(id))
            .get_result(&self.conn);

        match org_check {
            Ok(o) => Ok(o),
            Err(_e) => Err(eyre!("Could not retrieve org by id from DB")),
        }
    }

    pub fn secret_from_id(&self, id: i32) -> Option<Secret> {
        let secret_check: Result<Secret, _> = secret::table
            .select(secret::all_columns)
            .filter(secret::id.eq(id))
            .get_result(&self.conn);

        match secret_check {
            Ok(o) => Some(o),
            Err(_e) => None,
        }
    }

    pub fn repo_increment_build_index(&self, repo: Repo) -> Result<Repo> {
        //let _org_name = self.org_from_id(repo.org_id)?.name;

        let update_repo = NewRepo {
            org_id: repo.org_id,
            name: repo.name.clone(),
            uri: repo.uri,
            canonical_branch: repo.canonical_branch,
            git_host_type: repo.git_host_type,
            secret_id: repo.secret_id,
            build_active_state: repo.build_active_state,
            notify_active_state: repo.notify_active_state,
            next_build_index: repo.next_build_index + 1,
            remote_branch_heads: repo.remote_branch_heads,
        };

        let update_result = self.repo_update(&repo.name, update_repo)?;

        Ok(update_result.1)
    }

    pub fn org_add(&self, name: &str) -> Result<Org> {
        // Only insert if there are no other orgs by this name
        let org_check: Result<Org, _> = org::table
            .select(org::all_columns)
            .filter(org::name.eq(&name))
            .order_by(org::id)
            .get_result(&self.conn);

        match org_check {
            Err(_e) => {
                debug!("org doesn't exist. Inserting into db.");
                Ok(diesel::insert_into(org::table)
                    .values(NewOrg {
                        name: name.to_string(),
                        ..Default::default()
                    })
                    .get_result(&self.conn)
                    .expect("Error saving new org"))
            }
            Ok(o) => {
                debug!("org found in db. Returning result.");
                Ok(o)
            }
        }
    }

    pub fn org_get(&self, name: &str) -> Result<Org> {
        let mut org_check: Vec<Org> = org::table
            .select(org::all_columns)
            .filter(org::name.eq(&name))
            .order_by(org::id)
            .load(&self.conn)
            .expect("Error querying for org");

        match &org_check.len() {
            0 => {
                debug!("org not found in db");
                Err(eyre!("Org not Found"))
            }
            1 => {
                debug!("org found in db. Returning result.");
                Ok(org_check.pop().unwrap())
            }
            _ => Err(eyre!("Found more than one org by the same name in db")),
        }
    }

    pub fn org_update(&self, name: &str, update_org: NewOrg) -> Result<Org> {
        let org_update: Org = diesel::update(org::table)
            .filter(org::name.eq(&name))
            .set(update_org)
            .get_result(&self.conn)
            .expect("Error updating org");

        debug!("Result after update: {:?}", &org_update);

        Ok(org_update)
    }

    pub fn org_remove(&self, name: &str) -> Result<Org> {
        let org_delete: Org = diesel::delete(org::table)
            .filter(org::name.eq(&name))
            .get_result(&self.conn)
            .expect("Error deleting org");

        Ok(org_delete)
    }

    pub fn org_list(&self) -> Result<Vec<Org>> {
        let query: Vec<Org> = org::table
            .select(org::all_columns)
            .order_by(org::id)
            .load(&self.conn)
            .expect("Error getting all order records");

        Ok(query)
    }

    pub fn secret_add(&self, name: &str, secret_type: SecretType) -> Result<(Secret, Org)> {
        if let Some(ref org) = self.org {
            let query_result: Result<(Secret, Org), _> = secret::table
                .inner_join(org::table)
                .select((secret::all_columns, org::all_columns))
                .filter(secret::name.eq(&name))
                .filter(org::name.eq(&org))
                .get_result(&self.conn);

            match query_result {
                Err(_e) => {
                    debug!("secret doesn't exist. Inserting into db.");

                    let org_db = self.org_get(org).expect("Unable to find org");

                    let secret_db = diesel::insert_into(secret::table)
                        .values(NewSecret {
                            name: name.to_string(),
                            org_id: org_db.id,
                            secret_type,
                            vault_path: orb_vault_path(
                                &org_db.name,
                                name,
                                format!("{:?}", &secret_type).as_str(),
                            ),
                            ..Default::default()
                        })
                        .get_result(&self.conn)
                        .expect("Error saving new secret");

                    Ok((secret_db, org_db))
                }
                Ok((secret_db, org_db)) => {
                    debug!("secret found in db. Returning result.");
                    Ok((secret_db, org_db))
                }
            }
        } else {
            Err(eyre!("Org not set"))
        }
    }

    pub fn secret_get(&self, name: &str, _secret_type: SecretType) -> Result<(Secret, Org)> {
        if let Some(ref org) = self.org {
            let query_result: (Secret, Org) = secret::table
                .inner_join(org::table)
                .select((secret::all_columns, org::all_columns))
                .filter(secret::name.eq(&name))
                .filter(org::name.eq(&org))
                .first(&self.conn)
                .expect("Error querying for secret");

            debug!("Secret get result: \n {:?}", &query_result);

            Ok(query_result)
        } else {
            Err(eyre!("Org not set"))
        }
    }

    pub fn secret_update(&self, name: &str, update_secret: NewSecret) -> Result<Secret> {
        if let Some(ref org) = self.org {
            let org_db = self.org_get(org).expect("Unable to find org");

            let secret_update: Secret = diesel::update(secret::table)
                .filter(secret::org_id.eq(&org_db.id))
                .filter(secret::name.eq(&name))
                .set(update_secret)
                .get_result(&self.conn)
                .expect("Error updating secret");

            debug!("Result after update: {:?}", &secret_update);

            Ok(secret_update)
        } else {
            Err(eyre!("Org not set"))
        }
    }

    pub fn secret_remove(&self, name: &str, _secret_type: SecretType) -> Result<Secret> {
        if let Some(ref org) = self.org {
            let org_db = self.org_get(org).expect("Unable to find org");

            let secret_delete: Secret = diesel::delete(secret::table)
                .filter(secret::org_id.eq(&org_db.id))
                .filter(secret::name.eq(&name))
                .get_result(&self.conn)
                .expect("Error deleting secret");

            Ok(secret_delete)
        } else {
            Err(eyre!("Org not set"))
        }
    }

    pub fn secret_list(&self, filter: Option<SecretType>) -> Result<Vec<(Secret, Org)>> {
        if let Some(ref org) = self.org {
            let query_result: Vec<(Secret, Org)> = match filter {
                None => secret::table
                    .inner_join(org::table)
                    .select((secret::all_columns, org::all_columns))
                    .filter(org::name.eq(&org))
                    .load(&self.conn)
                    .expect("Error getting all secret records"),
                Some(_f) => secret::table
            .inner_join(org::table)
            .select((secret::all_columns, org::all_columns))
            .filter(org::name.eq(&org))
            //.filter(secret::secret_type.eq(SecretType::from(f))) // Not working yet.
            .load(&self.conn)
            .expect("Error getting all secret records"),
            };

            debug!("Secret list result: \n {:?}", &query_result);

            Ok(query_result)
        } else {
            Err(eyre!("Org not set"))
        }
    }

    pub fn repo_add(
        &self,

        // OrbitalClient
        name: &str,
        uri: &str,
        canonical_branch: &str,
        secret: Option<Secret>,
        branches_latest: serde_json::Value,
    ) -> Result<(Org, Repo, Option<Secret>)> {
        if let Some(ref org) = self.org {
            if let Some(ref _repo) = self.repo {
                let repo_check = self.repo_get();

                match repo_check {
                    Err(_e) => {
                        debug!("repo doesn't exist. Inserting into db.");

                        let secret_id = secret.map(|s| s.id);

                        let org_db = self.org_get(org)?;

                        let result: Repo = diesel::insert_into(repo::table)
                            .values(NewRepo {
                                name: name.into(),
                                org_id: org_db.id,
                                uri: uri.into(),
                                canonical_branch: canonical_branch.into(),
                                secret_id,
                                remote_branch_heads: branches_latest,
                                ..Default::default()
                            })
                            .get_result(&self.conn)
                            .expect("Error saving new repo");

                        debug!("DB insert result: {:?}", &result);

                        // Run query again. This time it should pass
                        self.repo_get()
                    }
                    Ok((o, r, s)) => {
                        debug!("repo found in db. Returning result.");
                        Ok((o, r, s))
                    }
                }
            } else {
                Err(eyre!("Repo not set"))
            }
        } else {
            Err(eyre!("Org not set"))
        }
    }

    pub fn repo_get(&self) -> Result<(Org, Repo, Option<Secret>)> {
        if let Some(ref org) = self.org {
            if let Some(ref repo) = self.repo {
                debug!("Repo get: Org: {}, Name: {}", org, repo);

                let query: Result<(Org, Repo), _> = repo::table
        .inner_join(org::table)
        .select((org::all_columns, repo::all_columns))
        .filter(repo::name.eq(&repo))
        //.filter(secret::id.eq(&secret_id))
        .get_result(&self.conn);

                match query {
                    Ok((o, r)) => {
                        // If we're using a secret, we should also return it to the requester
                        let secret = match &r.secret_id {
                            None => None,
                            Some(id) => self.secret_from_id(*id),
                        };

                        Ok((o, r, secret))
                    }
                    Err(_e) => Err(eyre!("{} not found in {} org", repo, org)),
                }
            } else {
                Err(eyre!("No repo set"))
            }
        } else {
            Err(eyre!("No org set"))
        }
    }

    // You should update your secret with secret_update()
    pub fn repo_update(
        &self,
        name: &str,
        update_repo: NewRepo,
    ) -> Result<(Org, Repo, Option<Secret>)> {
        if let Some(ref _org) = self.org {
            if let Some(ref _repo) = self.repo {
                let (org_db, _repo_db, secret_db) = self.repo_get()?;

                debug!("Right before updating DB row: {:?}", &update_repo);

                let repo_update: Repo = diesel::update(repo::table)
                    .filter(repo::org_id.eq(&org_db.id))
                    .filter(repo::name.eq(&name))
                    .set(update_repo)
                    .get_result(&self.conn)
                    .expect("Error updating repo");

                debug!("Result after update: {:?}", &repo_update);

                Ok((org_db, repo_update, secret_db))
            } else {
                Err(eyre!("No repo set"))
            }
        } else {
            Err(eyre!("No org set"))
        }
    }

    pub fn repo_remove(&self, name: &str) -> Result<(Org, Repo, Option<Secret>)> {
        if let Some(ref _repo) = self.repo {
            let (org_db, repo_db, secret_db) = self.repo_get()?;

            let _repo_delete: Repo = diesel::delete(repo::table)
                .filter(repo::org_id.eq(&org_db.id))
                .filter(repo::name.eq(&name))
                .get_result(&self.conn)
                .expect("Error deleting repo");

            Ok((org_db, repo_db, secret_db))
        } else {
            Err(eyre!("No repo set"))
        }
    }

    pub fn repo_list(&self) -> Result<Vec<(Org, Repo, Option<Secret>)>> {
        if let Some(ref org) = self.org {
            let query: Vec<(Org, Repo)> = repo::table
                .inner_join(org::table)
                .select((org::all_columns, repo::all_columns))
                .filter(org::name.eq(org))
                .load(&self.conn)
                .expect("Error selecting all repo");

            let map_result: Vec<(Org, Repo, Option<Secret>)> = query
                .into_iter()
                .map(|(o, r)| match r.clone().secret_id {
                    None => (o, r, None),
                    Some(id) => (o, r, self.secret_from_id(id)),
                })
                .collect();

            Ok(map_result)
        } else {
            Err(eyre!("Org not set"))
        }
    }

    pub fn build_target_add(
        &self,
        hash: &str,
        branch: &str,
        user_envs: Option<String>,
        job_trigger: JobTrigger,
    ) -> Result<(Org, Repo, BuildTarget)> {
        if let Some(ref _repo) = self.repo {
            let (org_db, repo_db, _) = self.repo_get()?;

            let build_target = NewBuildTarget {
                repo_id: repo_db.id,
                git_hash: hash.to_string(),
                branch: branch.to_string(),
                user_envs,
                build_index: repo_db.next_build_index,
                trigger: job_trigger,
                ..Default::default()
            };

            debug!("Build spec to insert: {:?}", &build_target);

            let result: BuildTarget = diesel::insert_into(build_target::table)
                .values(build_target)
                .get_result(&self.conn)
                .expect("Error saving new build_target");

            // Increment repo next_build_target by 1
            let updated_repo = self.repo_increment_build_index(repo_db)?;

            Ok((org_db, updated_repo, result))
        } else {
            Err(eyre!("Repo not set"))
        }
    }

    // This should probably return a Vec
    // Consider taking a Repo as input
    pub fn build_target_get(
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
    ) -> Result<(Org, Repo, Option<BuildTarget>)> {
        debug!(
        "Build target get request: org {:?} repo: {:?} hash: {:?} branch: {:?} build_index: {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index,
    );

        if let Some(ref _repo) = self.repo {
            let (org_db, repo_db, _secret_db) = self.repo_get()?;

            let result: Result<(Repo, BuildTarget), _> = build_target::table
                .inner_join(repo::table)
                .select((repo::all_columns, build_target::all_columns))
                .filter(build_target::repo_id.eq(repo_db.id))
                .filter(build_target::branch.eq(branch))
                .filter(build_target::build_index.eq(build_index))
                .get_result(&self.conn);

            match result {
                Ok((repo, build_target)) => {
                    debug!("BuildTarget found: {:?}", &build_target);
                    Ok((org_db, repo, Some(build_target)))
                }
                Err(_e) => Ok((org_db, repo_db, None)),
            }
        } else {
            Err(eyre!("No repo set"))
        }
    }

    //pub fn build_target_update(
    //    &self,
    //    org: &str,
    //    repo: &str,
    //    hash: &str,
    //    branch: &str,
    //    build_index: i32,
    //    update_build_target: NewBuildTarget,
    //) -> Result<(Org, Repo, BuildTarget)> {
    //    let (org_db, repo_db, build_target_db_opt) =
    //        build_target_get(conn, org, repo, hash, branch, build_index)?;
    //
    //    let build_target_db = build_target_db_opt.expect("No build target found");
    //
    //    let result: BuildTarget = diesel::update(build_target::table)
    //        .filter(build_target::id.eq(build_target_db.id))
    //        .set(update_build_target)
    //        .get_result(&self.conn)
    //        .expect("Error updating build target");
    //
    //    Ok((org_db, repo_db, result))
    //}

    //pub fn build_target_remove() {
    //    unimplemented!();
    //}
    //
    //pub fn build_target_list(
    //    &self,
    //    org: &str,
    //    repo: &str,
    //    limit: i32,
    //) -> Result<Vec<(Org, Repo, BuildTarget)>> {
    //    debug!(
    //        "Build target list request: org {:?} repo: {:?} limit: {:?}",
    //        &org, &repo, &limit
    //    );
    //
    //    let (org_db, _repo_db, _secret_db) = repo_get(conn, org, repo)?;
    //
    //    let result: Vec<(Repo, BuildTarget)> = build_target::table
    //        .inner_join(repo::table)
    //        .select((repo::all_columns, build_target::all_columns))
    //        .limit(limit.into())
    //        .load(&self.conn)
    //        .expect("Error saving new build_target");
    //
    //    let map_result: Vec<(Org, Repo, BuildTarget)> = result
    //        .into_iter()
    //        .map(|(r, b)| (org_db.clone(), r, b))
    //        .collect();
    //
    //    Ok(map_result)
    //}

    pub fn build_summary_add(
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
        build_summary: NewBuildSummary,
    ) -> Result<(Repo, BuildTarget, BuildSummary)> {
        debug!(
        "Build summary add request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index,
    );

        let (_org_db, repo_db, build_target_db_opt) =
            self.build_target_get(hash, branch, build_index)?;

        let build_target_db = build_target_db_opt.expect("Build target not found");

        debug!("Build summary to insert: {:?}", &build_summary);

        let result: BuildSummary = diesel::insert_into(build_summary::table)
            .values(build_summary)
            .get_result(&self.conn)
            .expect("Error saving new build_summary");

        Ok((repo_db, build_target_db, result))
    }

    pub fn build_summary_get(
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
    ) -> Result<(Repo, BuildTarget, Option<BuildSummary>)> {
        debug!(
        "Build summary get request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index,
    );

        let (_org_db, repo_db, build_target_db_opt) =
            self.build_target_get(hash, branch, build_index)?;

        let build_target_db = build_target_db_opt.expect("No build target found");

        let result: Result<(BuildTarget, BuildSummary), _> = build_summary::table
            .inner_join(build_target::table)
            .select((build_target::all_columns, build_summary::all_columns))
            .filter(build_summary::build_target_id.eq(build_target_db.id))
            .get_result(&self.conn);

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
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
        update_summary: NewBuildSummary,
    ) -> Result<(Repo, BuildTarget, BuildSummary)> {
        debug!(
        "Build summary update request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} update_summary: {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index, &update_summary,
    );

        let (org_db, build_target_db, build_summary_db_opt) =
            self.build_summary_get(hash, branch, build_index)?;

        let build_summary_db = build_summary_db_opt.expect("No build summary found");

        let result: BuildSummary = diesel::update(build_summary::table)
            .filter(build_summary::id.eq(build_summary_db.id))
            .set(update_summary)
            .get_result(&self.conn)
            .expect("Error updating build summary");

        Ok((org_db, build_target_db, result))
    }

    //pub fn build_summary_remove() {
    //    unimplemented!();
    //}

    // TODO: `repo` should be changed to Option<&str> for granularity between all or one repo
    pub fn build_summary_list(&self, limit: i32) -> Result<Vec<(Repo, BuildTarget, BuildSummary)>> {
        debug!(
            "Build summary list request: org {:?} repo: {:?} limit: {:?}",
            &self.org, &self.repo, &limit
        );

        let (_org_db, repo_db, _secret_db) = self.repo_get()?;

        let result: Vec<(BuildTarget, BuildSummary)> = build_summary::table
            .inner_join(build_target::table)
            .select((build_target::all_columns, build_summary::all_columns))
            .filter(build_target::repo_id.eq(repo_db.id))
            .order(build_summary::id.desc())
            .limit(limit.into())
            .load(&self.conn)
            .expect("Error listing build summaries");

        let map_result: Vec<(Repo, BuildTarget, BuildSummary)> = result
            .into_iter()
            .map(|(build_target, build_summary)| (repo_db.clone(), build_target, build_summary))
            .collect();

        Ok(map_result)
    }

    pub fn build_stage_add(
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
        build_summary_id: i32,
        build_stage: NewBuildStage,
    ) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
        debug!(
        "Build stage add request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} build_summary_id {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index, &build_summary_id,
    );

        let (_org_db, build_target_db, build_summary_db_opt) =
            self.build_summary_get(hash, branch, build_index)?;

        let build_summary_db = build_summary_db_opt.expect("No build summary found");

        debug!("Build stage to insert: {:?}", &build_stage);

        let result: BuildStage = diesel::insert_into(build_stage::table)
            .values(build_stage)
            .get_result(&self.conn)
            .expect("Error saving new build_stage");

        Ok((build_target_db, build_summary_db, result))
    }

    pub fn build_stage_get(
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
        build_summary_id: i32,
        build_stage_id: i32,
    ) -> Result<(BuildTarget, BuildSummary, Option<BuildStage>)> {
        debug!(
        "Build stage get request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} build_summary_id {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index, &build_summary_id,
    );

        let (_repo_db, build_target_db, build_summary_db_opt) =
            self.build_summary_get(hash, branch, build_index)?;

        let build_summary_db = build_summary_db_opt.expect("No build target found");

        let result: Result<(BuildSummary, BuildStage), _> = build_stage::table
            .inner_join(build_summary::table)
            .select((build_summary::all_columns, build_stage::all_columns))
            .filter(build_summary::build_target_id.eq(build_target_db.id))
            .filter(build_stage::id.eq(build_stage_id))
            .get_result(&self.conn);

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
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
        build_summary_id: i32,
        build_stage_id: i32,
        update_stage: NewBuildStage,
    ) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
        debug!(
        "Build stage update request: org: {:?} repo {:?} hash: {:?} branch {:?} build_index: {:?} build_summary_id {:?} build_stage_id {:?} update_stage {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index, &build_summary_id, &build_stage_id, &update_stage,
    );

        let (build_target_db, build_summary_db, build_stage_db_opt) =
            self.build_stage_get(hash, branch, build_index, build_summary_id, build_stage_id)?;

        let _build_stage_db = build_stage_db_opt.expect("No build stage found");

        let result: BuildStage = diesel::update(build_stage::table)
            .filter(build_stage::id.eq(build_stage_id))
            .set(update_stage)
            .get_result(&self.conn)
            .expect("Error updating build stage");

        Ok((build_target_db, build_summary_db, result))
    }

    //pub fn build_stage_remove() {
    //    unimplemented!();
    //}

    pub fn build_stage_list(
        &self,
        hash: &str,
        branch: &str,
        build_index: i32,
        limit: i32,
    ) -> Result<Vec<(BuildTarget, BuildSummary, BuildStage)>> {
        debug!(
        "Build stage list request: org {:?} repo: {:?} hash {:?} branch {:?} build_index {:?} limit: {:?}",
        &self.org, &self.repo, &hash, &branch, &build_index, &limit
    );

        let (_repo_db, build_target_db, build_summary_db_opt) =
            self.build_summary_get(hash, branch, build_index)?;

        let build_summary_db = build_summary_db_opt.expect("No build summary found");

        let result: Vec<(BuildSummary, BuildStage)> = build_stage::table
            .inner_join(build_summary::table)
            .select((build_summary::all_columns, build_stage::all_columns))
            .filter(build_summary::build_target_id.eq(build_target_db.id))
            .filter(build_stage::build_summary_id.eq(build_summary_db.id))
            .order(build_stage::id.asc())
            .limit(limit.into())
            .load(&self.conn)
            .expect("Error listing build stages");

        debug!(
            "Found {} stages for build id {}",
            &result.len(),
            build_index
        );

        let map_result: Vec<(BuildTarget, BuildSummary, BuildStage)> = result
            .into_iter()
            .map(|(build_summary, build_stage)| {
                (build_target_db.clone(), build_summary, build_stage)
            })
            .collect();

        Ok(map_result)
    }

    pub fn build_logs_get(
        &self,
        hash: &str,
        branch: &str,
        build_index: Option<i32>,
    ) -> Result<Vec<(BuildTarget, BuildSummary, BuildStage)>> {
        let (_org_db, repo_db, _secret_db) = self.repo_get()?;

        match build_index {
            Some(n) => self.build_stage_list(hash, branch, n, 255),
            None => self.build_stage_list(hash, branch, repo_db.next_build_index - 1, 255),
        }
    }

    pub fn is_build_cancelled(&self, hash: &str, branch: &str, build_index: i32) -> Result<bool> {
        match self.build_summary_get(hash, branch, build_index) {
            Ok((_, _, Some(summary))) => match summary.build_state {
                JobState::Canceled => Ok(true),
                _ => Ok(false),
            },
            Ok((_, _, None)) => {
                // Build hasn't been queued yet
                Ok(false)
            }
            Err(_) => Err(eyre!("Could not retrieve build summary from DB")),
        }
    }
}
