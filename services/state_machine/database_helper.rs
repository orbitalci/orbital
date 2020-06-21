use crate::state_machine::build_context::BuildContext;
use anyhow::Result;
use orbital_database::postgres;
use orbital_database::postgres::build_stage::{BuildStage, NewBuildStage};
use orbital_database::postgres::build_summary::{BuildSummary, NewBuildSummary};
use orbital_database::postgres::build_target::BuildTarget;
use orbital_database::postgres::org::Org;
use orbital_database::postgres::repo::Repo;

pub struct DbHelper;

impl DbHelper {
    pub fn is_build_cancelled(build_context: &BuildContext) -> Result<bool> {
        let pg_conn = postgres::client::establish_connection();

        // TODO: Need to make "cancelled" the consistent spelling...
        postgres::client::is_build_canceled(
            &pg_conn,
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().unwrap_or_default(),
            &build_context.branch,
            build_context.id.clone().unwrap(),
        )
    }

    pub fn build_target_add(build_context: &BuildContext) -> Result<(Org, Repo, BuildTarget)> {
        let pg_conn = postgres::client::establish_connection();

        postgres::client::build_target_add(
            &pg_conn,
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().expect("No repo hash to target"),
            &build_context.branch,
            Some(build_context.user_envs.clone().unwrap_or_default().join("")),
            build_context.job_trigger.clone(),
        )
    }

    pub fn build_summary_add(
        build_context: &BuildContext,
        new_build_summary: NewBuildSummary,
    ) -> Result<(Repo, BuildTarget, BuildSummary)> {
        let pg_conn = postgres::client::establish_connection();

        postgres::client::build_summary_add(
            &pg_conn,
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.clone().unwrap(),
            new_build_summary,
        )
    }

    pub fn build_summary_update(
        build_context: &BuildContext,
        update_summary: NewBuildSummary,
    ) -> Result<(Repo, BuildTarget, BuildSummary)> {
        let pg_conn = postgres::client::establish_connection();

        postgres::client::build_summary_update(
            &pg_conn,
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.clone().unwrap(),
            update_summary,
        )
    }

    pub fn build_stage_add(
        build_context: &BuildContext,
        new_build_stage: NewBuildStage,
    ) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
        let pg_conn = postgres::client::establish_connection();

        postgres::client::build_stage_add(
            &pg_conn,
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.clone().unwrap(),
            build_context._db_build_summary_id,
            new_build_stage,
        )
    }

    pub fn build_stage_update(
        build_context: &BuildContext,
        update_build_stage: NewBuildStage,
    ) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
        let pg_conn = postgres::client::establish_connection();

        postgres::client::build_stage_update(
            &pg_conn,
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.clone().unwrap(),
            build_context._db_build_summary_id,
            build_context._db_build_cur_stage_id,
            update_build_stage,
        )
    }
}
