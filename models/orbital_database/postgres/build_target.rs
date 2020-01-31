use crate::postgres::repo::Repo;
use crate::postgres::schema::{build_target, JobTrigger};
use chrono::{NaiveDateTime, Utc};
use prost_types::Timestamp;
use std::time::SystemTime;

use orbital_headers::build_meta;

#[derive(Insertable, Debug, PartialEq, AsChangeset, Clone)]
//#[belongs_to(Repo)]
#[table_name = "build_target"]
pub struct NewBuildTarget {
    pub repo_id: i32,
    pub git_hash: String,
    pub branch: String,
    pub user_envs: Option<String>,
    pub queue_time: NaiveDateTime,
    pub build_index: i32,
    pub trigger: JobTrigger,
}

impl Default for NewBuildTarget {
    fn default() -> Self {
        NewBuildTarget {
            repo_id: 0,
            //name: "".into(),
            git_hash: "".into(),
            branch: "".into(),
            user_envs: None,
            queue_time: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            build_index: 0,
            trigger: JobTrigger::Unspecified,
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable, Associations, QueryableByName)]
#[belongs_to(Repo)]
#[table_name = "build_target"]
pub struct BuildTarget {
    pub id: i32,
    pub repo_id: i32,
    pub git_hash: String,
    pub branch: String,
    pub user_envs: Option<String>,
    pub queue_time: NaiveDateTime,
    pub build_index: i32,
    pub trigger: JobTrigger,
}

impl Default for BuildTarget {
    fn default() -> Self {
        BuildTarget {
            id: 0,
            repo_id: 0,
            //name: "".into(),
            git_hash: "".into(),
            branch: "".into(),
            user_envs: None,
            queue_time: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            build_index: 0,
            trigger: JobTrigger::Unspecified,
        }
    }
}

// Does not set org at all, repo will be repo_id by default
// Loses queue time information
impl From<BuildTarget> for build_meta::BuildTarget {
    fn from(build_target: BuildTarget) -> Self {
        let queue_time_proto = Timestamp {
            seconds: build_target.queue_time.timestamp(),
            nanos: 0,
        };


        build_meta::BuildTarget {
            git_repo: format!("{:?}", build_target.repo_id),
            branch: build_target.branch,
            user_envs: match build_target.user_envs {
                Some(e) => e,
                None => String::new(),
            },
            commit_hash: build_target.git_hash,
            id: build_target.id,
            //queue_time: queue_time_proto,
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
            user_envs: match build_target.user_envs.len() {
                0 => None,
                _ => Some(build_target.user_envs),
            },

            //queue_time: queue_timestamp,
            //build_index: build_target.build_index,
            //job_trigger: build_target
            ..Default::default()
        }
    }
}
