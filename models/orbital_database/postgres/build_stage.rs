use crate::postgres::schema::build_stage;
use chrono::{NaiveDateTime, Utc};
use orbital_headers::build_meta::BuildStage as ProtoBuildStage;

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
    pub exit_code: Option<i32>,
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

impl From<BuildStage> for ProtoBuildStage {
    fn from(build_stage: BuildStage) -> Self {
        let start_time = Some(prost_types::Timestamp {
            seconds: build_stage.start_time.timestamp(),
            nanos: build_stage.start_time.timestamp_subsec_nanos() as i32,
        });

        let end_time = match build_stage.end_time {
            Some(end_time) => Some(prost_types::Timestamp {
                seconds: end_time.timestamp(),
                nanos: end_time.timestamp_subsec_nanos() as i32,
            }),
            None => None,
        };

        ProtoBuildStage {
            id: build_stage.id,
            build_id: build_stage.build_summary_id,
            stage: build_stage.stage_name.unwrap_or("".to_string()),
            exit_code: build_stage.exit_code.unwrap_or(-1),
            status: 0,
            start_time: start_time,
            end_time: end_time,
            output: build_stage
                .output
                .unwrap_or(String::new())
                .as_bytes()
                .to_vec(),
        }
    }
}
