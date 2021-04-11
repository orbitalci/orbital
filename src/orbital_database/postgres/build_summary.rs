use crate::orbital_database::postgres::schema::{build_summary, JobState};
use chrono::{NaiveDateTime, Utc};

#[derive(Insertable, Debug, PartialEq, AsChangeset, Clone)]
#[table_name = "build_summary"]
pub struct NewBuildSummary {
    pub build_target_id: i32,
    pub start_time: Option<NaiveDateTime>,
    pub end_time: Option<NaiveDateTime>,
    pub build_state: JobState,
}

impl Default for NewBuildSummary {
    fn default() -> Self {
        NewBuildSummary {
            build_target_id: 0,
            start_time: Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0)),
            end_time: None,
            build_state: JobState::Queued,
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable, Associations, QueryableByName)]
#[table_name = "build_summary"]
pub struct BuildSummary {
    pub id: i32,
    pub build_target_id: i32,
    pub start_time: Option<NaiveDateTime>,
    pub end_time: Option<NaiveDateTime>,
    pub build_state: JobState,
}

impl Default for BuildSummary {
    fn default() -> Self {
        BuildSummary {
            id: 0,
            build_target_id: 0,
            start_time: Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0)),
            end_time: None,
            build_state: JobState::Queued,
        }
    }
}
