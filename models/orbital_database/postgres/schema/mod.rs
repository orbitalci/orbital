// With our custom PG enums, we cannot allow diesel-cli to modify this file.
// See: diesel-rs/diesel#343
use diesel::deserialize::{self, FromSql};
use diesel::pg::Pg;
use diesel::serialize::{self, IsNull, Output, ToSql};
use std::io::Write;
use strum_macros::{Display, EnumString, EnumVariantNames};

use orbital_headers::orbital_types;

// Custom type handling reference:
// https://github.com/diesel-rs/diesel/blob/1.4.x/diesel_tests/tests/custom_types.rs

table! {
    use diesel::sql_types::{Integer, Text, Timestamp};
    use super::ActiveStatePGEnum;

    org (id) {
        id -> Integer,
        name -> Text,
        created -> Timestamp,
        last_update -> Timestamp,
        active_state -> ActiveStatePGEnum,
    }
}

#[derive(SqlType, Debug)]
#[postgres(type_name = "active_state")]
pub struct ActiveStatePGEnum;

#[derive(
    Debug, Clone, PartialEq, FromSqlRow, AsExpression, EnumString, EnumVariantNames, Display,
)]
#[strum(serialize_all = "snake_case")]
#[sql_type = "ActiveStatePGEnum"]
pub enum ActiveState {
    Unspecified = 0,
    Unknown = 1,
    Enabled = 2,
    Disabled = 3,
    Deleted = 4,
}

impl From<i32> for ActiveState {
    fn from(active_state: i32) -> Self {
        match active_state {
            0 => ActiveState::Unspecified,
            1 => ActiveState::Unknown,
            2 => ActiveState::Enabled,
            3 => ActiveState::Disabled,
            4 => ActiveState::Deleted,
            _ => panic!("Unrecognized ActiveState variant"),
        }
    }
}

impl From<ActiveState> for i32 {
    fn from(active_state: ActiveState) -> Self {
        match active_state {
            ActiveState::Unspecified => 0,
            ActiveState::Unknown => 1,
            ActiveState::Enabled => 2,
            ActiveState::Disabled => 3,
            ActiveState::Deleted => 4,
        }
    }
}

impl ToSql<ActiveStatePGEnum, Pg> for ActiveState {
    fn to_sql<W: Write>(&self, out: &mut Output<W, Pg>) -> serialize::Result {
        match *self {
            ActiveState::Unspecified => out.write_all(b"")?,
            ActiveState::Unknown => out.write_all(b"unknown")?,
            ActiveState::Enabled => out.write_all(b"enabled")?,
            ActiveState::Disabled => out.write_all(b"disabled")?,
            ActiveState::Deleted => out.write_all(b"deleted")?,
        }
        Ok(IsNull::No)
    }
}

impl FromSql<ActiveStatePGEnum, Pg> for ActiveState {
    fn from_sql(bytes: Option<&[u8]>) -> deserialize::Result<Self> {
        match not_none!(bytes) {
            b"" => Ok(ActiveState::Unspecified),
            b"unknown" => Ok(ActiveState::Unknown),
            b"enabled" => Ok(ActiveState::Enabled),
            b"disabled" => Ok(ActiveState::Disabled),
            b"deleted" => Ok(ActiveState::Deleted),
            _ => Err("Unrecognized ActiveState variant".into()),
        }
    }
}

// Convert from the proto codegen structs to the diesel structs
impl From<orbital_types::ActiveState> for ActiveState {
    fn from(active_state: orbital_types::ActiveState) -> Self {
        match active_state {
            orbital_types::ActiveState::Unspecified => ActiveState::Unspecified,
            orbital_types::ActiveState::Unknown => ActiveState::Unknown,
            orbital_types::ActiveState::Enabled => ActiveState::Enabled,
            orbital_types::ActiveState::Disabled => ActiveState::Disabled,
            orbital_types::ActiveState::Deleted => ActiveState::Deleted,
        }
    }
}

table! {
    use diesel::sql_types::{Integer, Text};
    use super::{ActiveStatePGEnum,SecretTypePGEnum};

    secret (id) {
        id -> Integer,
        org_id -> Integer,
        name -> Text,
        secret_type -> SecretTypePGEnum,
        vault_path -> Text,
        active_state -> ActiveStatePGEnum,
    }
}

joinable!(secret -> org (org_id));

#[derive(SqlType, Debug, Queryable)]
#[postgres(type_name = "secret_type")]
pub struct SecretTypePGEnum;

#[derive(
    Debug, Clone, Copy, PartialEq, FromSqlRow, AsExpression, EnumString, EnumVariantNames, Display,
)]
#[strum(serialize_all = "snake_case")]
#[sql_type = "SecretTypePGEnum"]
pub enum SecretType {
    Unspecified = 0,
    BasicAuth = 1,
    ApiKey = 2,
    EnvVar = 3,
    File = 4,
    SshKey = 5,
    DockerRegistry = 6,
    NpmRepo = 7,
    PypiRegistry = 8,
    MavenRepo = 9,
    Kubernetes = 10,
}

impl From<i32> for SecretType {
    fn from(secret_type: i32) -> Self {
        match secret_type {
            0 => SecretType::Unspecified,
            1 => SecretType::BasicAuth,
            2 => SecretType::ApiKey,
            3 => SecretType::EnvVar,
            4 => SecretType::File,
            5 => SecretType::SshKey,
            6 => SecretType::DockerRegistry,
            7 => SecretType::NpmRepo,
            8 => SecretType::PypiRegistry,
            9 => SecretType::MavenRepo,
            10 => SecretType::Kubernetes,
            _ => panic!("Unrecognized SecretType variant"),
        }
    }
}

impl From<SecretType> for i32 {
    fn from(secret_type: SecretType) -> Self {
        match secret_type {
            SecretType::Unspecified => 0,
            SecretType::BasicAuth => 1,
            SecretType::ApiKey => 2,
            SecretType::EnvVar => 3,
            SecretType::File => 4,
            SecretType::SshKey => 5,
            SecretType::DockerRegistry => 6,
            SecretType::NpmRepo => 7,
            SecretType::PypiRegistry => 8,
            SecretType::MavenRepo => 9,
            SecretType::Kubernetes => 10,
        }
    }
}

impl ToSql<SecretTypePGEnum, Pg> for SecretType {
    fn to_sql<W: Write>(&self, out: &mut Output<W, Pg>) -> serialize::Result {
        match *self {
            SecretType::Unspecified => out.write_all(b"")?,
            SecretType::BasicAuth => out.write_all(b"basic_auth")?,
            SecretType::ApiKey => out.write_all(b"api_key")?,
            SecretType::EnvVar => out.write_all(b"env_var")?,
            SecretType::File => out.write_all(b"file")?,
            SecretType::SshKey => out.write_all(b"ssh_key")?,
            SecretType::DockerRegistry => out.write_all(b"docker_registry")?,
            SecretType::NpmRepo => out.write_all(b"npm_repo")?,
            SecretType::PypiRegistry => out.write_all(b"pypi_registry")?,
            SecretType::MavenRepo => out.write_all(b"maven_repo")?,
            SecretType::Kubernetes => out.write_all(b"kubernetes")?,
        }
        Ok(IsNull::No)
    }
}

impl FromSql<SecretTypePGEnum, Pg> for SecretType {
    fn from_sql(bytes: Option<&[u8]>) -> deserialize::Result<Self> {
        match not_none!(bytes) {
            b"" => Ok(SecretType::Unspecified),
            b"basic_auth" => Ok(SecretType::BasicAuth),
            b"api_key" => Ok(SecretType::ApiKey),
            b"env_var" => Ok(SecretType::EnvVar),
            b"file" => Ok(SecretType::File),
            b"ssh_key" => Ok(SecretType::SshKey),
            b"docker_registry" => Ok(SecretType::DockerRegistry),
            b"npm_repo" => Ok(SecretType::NpmRepo),
            b"pypi_registry" => Ok(SecretType::PypiRegistry),
            b"maven_repo" => Ok(SecretType::MavenRepo),
            b"kubernetes" => Ok(SecretType::Kubernetes),
            _ => Err("Unrecognized SecretType variant".into()),
        }
    }
}

// Convert from the proto codegen structs to the diesel structs
impl From<orbital_types::SecretType> for SecretType {
    fn from(secret_type: orbital_types::SecretType) -> Self {
        match secret_type {
            orbital_types::SecretType::Unspecified => SecretType::Unspecified,
            orbital_types::SecretType::BasicAuth => SecretType::BasicAuth,
            orbital_types::SecretType::ApiKey => SecretType::ApiKey,
            orbital_types::SecretType::EnvVar => SecretType::EnvVar,
            orbital_types::SecretType::File => SecretType::File,
            orbital_types::SecretType::SshKey => SecretType::SshKey,
            orbital_types::SecretType::DockerRegistry => SecretType::DockerRegistry,
            orbital_types::SecretType::NpmRepo => SecretType::NpmRepo,
            orbital_types::SecretType::PypiRegistry => SecretType::PypiRegistry,
            orbital_types::SecretType::MavenRepo => SecretType::MavenRepo,
            orbital_types::SecretType::Kubernetes => SecretType::Kubernetes,
        }
    }
}

// FIXME: This is missing a column for storing ad-hoc env vars
table! {
    use diesel::sql_types::{Integer, Text, Nullable};
    use super::{ActiveStatePGEnum,GitHostTypePGEnum};

    repo (id) {
        id -> Integer,
        org_id -> Integer,
        name -> Text,
        uri -> Text,
        git_host_type -> GitHostTypePGEnum,
        secret_id -> Nullable<Integer>,
        build_active_state -> ActiveStatePGEnum,
        notify_active_state -> ActiveStatePGEnum,
        next_build_index -> Integer,
    }
}

joinable!(repo -> org(org_id));
joinable!(repo -> secret(secret_id));

allow_tables_to_appear_in_same_query!(org, secret, repo);

#[derive(SqlType, Debug)]
#[postgres(type_name = "git_host_type")]
pub struct GitHostTypePGEnum;

#[derive(
    Debug, Clone, Copy, PartialEq, FromSqlRow, AsExpression, EnumString, EnumVariantNames, Display,
)]
#[sql_type = "GitHostTypePGEnum"]
pub enum GitHostType {
    Unspecified = 0,
    Generic = 1,
    Bitbucket = 2,
    Github = 3,
}

impl ToSql<GitHostTypePGEnum, Pg> for GitHostType {
    fn to_sql<W: Write>(&self, out: &mut Output<W, Pg>) -> serialize::Result {
        match *self {
            GitHostType::Unspecified => out.write_all(b"")?,
            GitHostType::Generic => out.write_all(b"generic")?,
            GitHostType::Bitbucket => out.write_all(b"bitbucket")?,
            GitHostType::Github => out.write_all(b"github")?,
        }
        Ok(IsNull::No)
    }
}

impl FromSql<GitHostTypePGEnum, Pg> for GitHostType {
    fn from_sql(bytes: Option<&[u8]>) -> deserialize::Result<Self> {
        match not_none!(bytes) {
            b"" => Ok(GitHostType::Unspecified),
            b"generic" => Ok(GitHostType::Generic),
            b"bitbucket" => Ok(GitHostType::Bitbucket),
            b"github" => Ok(GitHostType::Github),
            _ => Err("Unrecognized GitHostType variant".into()),
        }
    }
}

table! {
    use diesel::sql_types::{Integer, Text, Nullable, Timestamp};
    use super::JobTriggerPGEnum;

    build_target (id) {
        id -> Integer,
        repo_id -> Integer,
        git_hash -> Text,
        branch -> Text,
        user_envs -> Nullable<Text>,
        queue_time -> Timestamp,
        build_index -> Integer,
        trigger -> JobTriggerPGEnum,
    }
}

joinable!(build_target -> repo(repo_id));
allow_tables_to_appear_in_same_query!(build_target, repo);

#[derive(SqlType, Debug)]
#[postgres(type_name = "job_trigger")]
pub struct JobTriggerPGEnum;

#[derive(
    Debug, Clone, Copy, PartialEq, FromSqlRow, AsExpression, EnumString, EnumVariantNames, Display,
)]
#[sql_type = "JobTriggerPGEnum"]
pub enum JobTrigger {
    Unspecified = 0,
    Push = 1,
    PullRequest = 2,
    Webhook = 3,
    Poll = 4,
    Manual = 5,
    SubscribeTrigger = 6,
    CommitMsgTrigger = 7,
}

impl From<i32> for JobTrigger {
    fn from(job_trigger: i32) -> Self {
        match job_trigger {
            0 => JobTrigger::Unspecified,
            1 => JobTrigger::Push,
            2 => JobTrigger::PullRequest,
            3 => JobTrigger::Webhook,
            4 => JobTrigger::Poll,
            5 => JobTrigger::Manual,
            6 => JobTrigger::SubscribeTrigger,
            7 => JobTrigger::CommitMsgTrigger,
            _ => panic!("Unrecognized JobTrigger variant"),
        }
    }
}

impl From<JobTrigger> for i32 {
    fn from(job_trigger: JobTrigger) -> Self {
        match job_trigger {
            JobTrigger::Unspecified => 0,
            JobTrigger::Push => 1,
            JobTrigger::PullRequest => 2,
            JobTrigger::Webhook => 3,
            JobTrigger::Poll => 4,
            JobTrigger::Manual => 5,
            JobTrigger::SubscribeTrigger => 6,
            JobTrigger::CommitMsgTrigger => 7,
        }
    }
}

impl ToSql<JobTriggerPGEnum, Pg> for JobTrigger {
    fn to_sql<W: Write>(&self, out: &mut Output<W, Pg>) -> serialize::Result {
        match *self {
            JobTrigger::Unspecified => out.write_all(b"")?,
            JobTrigger::Push => out.write_all(b"push")?,
            JobTrigger::PullRequest => out.write_all(b"pull_request")?,
            JobTrigger::Webhook => out.write_all(b"webhook")?,
            JobTrigger::Poll => out.write_all(b"poll")?,
            JobTrigger::Manual => out.write_all(b"manual")?,
            JobTrigger::SubscribeTrigger => out.write_all(b"subscribe_trigger")?,
            JobTrigger::CommitMsgTrigger => out.write_all(b"commit_msg_trigger")?,
        }
        Ok(IsNull::No)
    }
}

impl FromSql<JobTriggerPGEnum, Pg> for JobTrigger {
    fn from_sql(bytes: Option<&[u8]>) -> deserialize::Result<Self> {
        match not_none!(bytes) {
            b"" => Ok(JobTrigger::Unspecified),
            b"push" => Ok(JobTrigger::Push),
            b"pull_request" => Ok(JobTrigger::PullRequest),
            b"webhook" => Ok(JobTrigger::Webhook),
            b"poll" => Ok(JobTrigger::Poll),
            b"manual" => Ok(JobTrigger::Manual),
            b"subscribe_trigger" => Ok(JobTrigger::SubscribeTrigger),
            b"commit_msg_trigger" => Ok(JobTrigger::CommitMsgTrigger),
            _ => Err("Unrecognized GitHostType variant".into()),
        }
    }
}

table! {
    use diesel::sql_types::{Integer, Nullable, Timestamp};
    use super::JobStatePGEnum;

    build_summary (id) {
        id -> Integer,
        build_target_id -> Integer,
        start_time -> Nullable<Timestamp>,
        end_time -> Nullable<Timestamp>,
        build_state -> JobStatePGEnum,
    }
}

joinable!(build_summary -> build_target(build_target_id));
allow_tables_to_appear_in_same_query!(build_summary, build_target);

#[derive(SqlType, Debug)]
#[postgres(type_name = "job_state")]
pub struct JobStatePGEnum;

#[derive(
    Debug, Clone, Copy, PartialEq, FromSqlRow, AsExpression, EnumString, EnumVariantNames, Display,
)]
#[sql_type = "JobStatePGEnum"]
pub enum JobState {
    Unspecified = 0,
    Unknown = 1,
    Queued = 2,
    Starting = 3,
    Running = 4,
    Finishing = 5,
    Canceled = 6,
    SystemErr = 7,
    Failed = 8,
    Done = 9,
    Deleted = 10,
}

table! {
    use diesel::sql_types::{Integer, Text, Nullable, Timestamp};

    build_stage (id) {
        id -> Integer,
        build_summary_id -> Integer,
        build_host -> Nullable<Text>,
        stage_name -> Nullable<Text>,
        output -> Nullable<Text>,
        start_time -> Timestamp,
        end_time -> Nullable<Timestamp>,
        exit_code -> Nullable<Integer>,
    }
}

joinable!(build_stage -> build_summary(build_summary_id));
allow_tables_to_appear_in_same_query!(build_stage, build_summary);

impl ToSql<JobStatePGEnum, Pg> for JobState {
    fn to_sql<W: Write>(&self, out: &mut Output<W, Pg>) -> serialize::Result {
        match *self {
            JobState::Unspecified => out.write_all(b"")?,
            JobState::Unknown => out.write_all(b"unknown")?,
            JobState::Queued => out.write_all(b"queued")?,
            JobState::Starting => out.write_all(b"starting")?,
            JobState::Running => out.write_all(b"running")?,
            JobState::Finishing => out.write_all(b"finishing")?,
            JobState::Canceled => out.write_all(b"canceled")?,
            JobState::SystemErr => out.write_all(b"systemerr")?,
            JobState::Failed => out.write_all(b"failed")?,
            JobState::Done => out.write_all(b"done")?,
            JobState::Deleted => out.write_all(b"deleted")?,
        }
        Ok(IsNull::No)
    }
}

impl FromSql<JobStatePGEnum, Pg> for JobState {
    fn from_sql(bytes: Option<&[u8]>) -> deserialize::Result<Self> {
        match not_none!(bytes) {
            b"" => Ok(JobState::Unspecified),
            b"unknown" => Ok(JobState::Unknown),
            b"queued" => Ok(JobState::Queued),
            b"starting" => Ok(JobState::Starting),
            b"running" => Ok(JobState::Running),
            b"finishing" => Ok(JobState::Finishing),
            b"cancele" => Ok(JobState::Canceled),
            b"systemerr" => Ok(JobState::SystemErr),
            b"failed" => Ok(JobState::Failed),
            b"done" => Ok(JobState::Done),
            b"deleted" => Ok(JobState::Deleted),

            _ => Err("Unrecognized JobState variant".into()),
        }
    }
}
