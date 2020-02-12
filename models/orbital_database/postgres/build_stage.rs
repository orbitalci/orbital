use crate::postgres::schema::{build_stage};
use chrono::{NaiveDateTime, Utc};

// TODO: Stage name required? Build service can generate an index
#[derive(Insertable, Debug, PartialEq, AsChangeset, Clone)]
#[table_name = "build_stage"]
pub struct NewBuildStage {
    pub build_summary_id: i32,
    pub build_host: Option<String>,
    pub stage_name: Option<String>,
    pub output: Option<String>,
    pub start_time: NaiveDateTime,
    pub end_time: Option<NaiveDateTime>,
    pub exit_code: Option<i32>,
}

impl Default for NewBuildStage {
    fn default() -> Self {
        NewBuildStage {
            build_summary_id: 0,
            build_host: None,
            stage_name: None,
            output: None,
            start_time: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            end_time: None,
            exit_code: None,
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable, Associations, QueryableByName)]
#[table_name = "build_stage"]
pub struct BuildStage {
    pub id: i32,
    pub build_summary_id: i32,
    pub build_host: Option<String>,
    pub stage_name: Option<String>,
    pub output: Option<String>,
    pub start_time: NaiveDateTime,
    pub end_time: Option<NaiveDateTime>,
    pub exit_code: Option<i32>
}

impl Default for BuildStage {
    fn default() -> Self {
        BuildStage {
            id: 0,
            build_summary_id: 0,
            build_host: None,
            stage_name: None,
            output: None,
            start_time: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            end_time: None,
            exit_code: None,
        }
    }
}