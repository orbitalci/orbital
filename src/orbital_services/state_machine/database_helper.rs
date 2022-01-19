use super::build_context::BuildContext;
use crate::orbital_database::postgres;
use crate::orbital_database::postgres::build_stage::{BuildStage, NewBuildStage};
use crate::orbital_database::postgres::build_summary::{BuildSummary, NewBuildSummary};
use crate::orbital_database::postgres::build_target::BuildTarget;
use crate::orbital_database::postgres::org::Org;
use crate::orbital_database::postgres::repo::Repo;
use color_eyre::eyre::Result;

pub struct DbHelper;

impl DbHelper {
    pub fn is_build_cancelled(build_context: &BuildContext) -> Result<bool> {
        let orb_db = postgres::client::OrbitalDBClient::new();

        // TODO: Need to make "cancelled" the consistent spelling...
        orb_db.is_build_cancelled(
            &build_context.repo_name,
            &build_context.hash.clone().unwrap_or_default(),
            &build_context.branch,
            build_context.id.unwrap(),
        )
    }

    pub fn build_target_add(build_context: &BuildContext) -> Result<(Org, Repo, BuildTarget)> {
        let orb_db = postgres::client::OrbitalDBClient::new();

        orb_db.build_target_add(
            &build_context.repo_name,
            &build_context.hash.clone().expect("No repo hash to target"),
            &build_context.branch,
            Some(build_context.user_envs.clone().unwrap_or_default().join("")),
            build_context.job_trigger,
        )
    }

    pub fn build_summary_add(
        build_context: &BuildContext,
        new_build_summary: NewBuildSummary,
    ) -> Result<(Repo, BuildTarget, BuildSummary)> {
        let orb_db = postgres::client::OrbitalDBClient::new();

        orb_db.build_summary_add(
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.unwrap(),
            new_build_summary,
        )
    }

    pub fn build_summary_update(
        build_context: &BuildContext,
        update_summary: NewBuildSummary,
    ) -> Result<(Repo, BuildTarget, BuildSummary)> {
        let orb_db = postgres::client::OrbitalDBClient::new();

        orb_db.build_summary_update(
            &build_context.org,
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.unwrap(),
            update_summary,
        )
    }

    pub fn build_stage_add(
        build_context: &BuildContext,
        new_build_stage: NewBuildStage,
    ) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
        let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(build_context.org.to_string()));

        orb_db.build_stage_add(
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.unwrap(),
            build_context._db_build_summary_id,
            new_build_stage,
        )
    }

    pub fn build_stage_update(
        build_context: &BuildContext,
        update_build_stage: NewBuildStage,
    ) -> Result<(BuildTarget, BuildSummary, BuildStage)> {
        let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(build_context.org.to_string()));

        orb_db.build_stage_update(
            &build_context.repo_name,
            &build_context.hash.clone().unwrap(),
            &build_context.branch,
            build_context.id.unwrap(),
            build_context._db_build_summary_id,
            build_context._db_build_cur_stage_id,
            update_build_stage,
        )
    }
}
