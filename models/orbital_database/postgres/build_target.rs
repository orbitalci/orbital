use crate::postgres::repo::Repo;
use crate::postgres::schema::{build_target, JobTrigger};
use chrono::{NaiveDateTime, Utc};

use orbital_headers::build_meta;

#[derive(Insertable, Debug, PartialEq, AsChangeset)]
//#[belongs_to(Repo)]
#[table_name = "build_target"]
pub struct NewBuildTarget {
    pub repo_id: i32,
    pub name: String,
    pub git_hash: String,
    pub branch: String,
    pub build_index: i32,
    pub job_trigger: JobTrigger,
}

impl Default for NewBuildTarget {
    fn default() -> Self {
        NewBuildTarget {
            repo_id: 0,
            name: "".into(),
            git_hash: "".into(),
            branch: "".into(),
            build_index: 0,
            job_trigger: JobTrigger::Unspecified,
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable, Associations, QueryableByName)]
#[belongs_to(Repo)]
#[table_name = "build_target"]
pub struct BuildTarget {
    pub id: i32,
    pub repo_id: i32,
    pub name: String,
    pub git_hash: String,
    pub branch: String,
    pub queue_time: NaiveDateTime,
    pub build_index: i32,
    pub job_trigger: JobTrigger,
}

impl Default for BuildTarget {
    fn default() -> Self {
        BuildTarget {
            id: 0,
            repo_id: 0,
            name: "".into(),
            git_hash: "".into(),
            branch: "".into(),
            queue_time: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            build_index: 0,
            job_trigger: JobTrigger::Unspecified,
        }
    }
}

// Does not set org at all, repo will be repo_id by default
impl From<BuildTarget> for build_meta::BuildTarget {
    fn from(build_target: BuildTarget) -> Self {
        build_meta::BuildTarget {
            git_repo: format!("{:?}", build_target.repo_id),
            branch: build_target.branch,
            commit_hash: build_target.git_hash,
            id: build_target.id,
            ..Default::default()
        }
    }
}

impl From<build_meta::BuildTarget> for BuildTarget {
    fn from(build_target: build_meta::BuildTarget) -> Self {

        BuildTarget {
            id: build_target.id,
            repo_id: build_target.id,
            //name: build_target.name,
            git_hash: build_target.commit_hash,
            branch: build_target.branch,
            //queue_time: queue_timestamp,
            //build_index: build_target.build_index,
            //job_trigger: build_target
            ..Default::default()
        }
    }
}
