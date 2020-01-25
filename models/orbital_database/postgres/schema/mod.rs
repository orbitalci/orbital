// With our custom PG enums, we cannot allow diesel-cli to modify this file.
// See: diesel-rs/diesel#343
use diesel::deserialize::{self, FromSql};
use diesel::pg::Pg;
use diesel::serialize::{self, IsNull, Output, ToSql};
use std::io::Write;
use std::str::FromStr;

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

#[derive(Debug, Clone, PartialEq, FromSqlRow, AsExpression)]
#[sql_type = "ActiveStatePGEnum"]
pub enum ActiveState {
    Unspecified = 0,
    Unknown = 1,
    Enabled = 2,
    Disabled = 3,
    Deleted = 4,
}

impl ActiveState {
    /// A list of possible variants in `&'static str` form
    pub fn variants() -> [&'static str; 5] {
        ["unspecified", "unknown", "enabled", "disabled", "deleted"]
    }
}

impl FromStr for ActiveState {
    type Err = std::string::ParseError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Ok(match s.to_lowercase().as_ref() {
            "unspecified" => ActiveState::Unspecified,
            "unknown" => ActiveState::Unknown,
            "enabled" => ActiveState::Enabled,
            "disabled" => ActiveState::Disabled,
            "deleted" => ActiveState::Deleted,
            _ => ActiveState::Unknown,
        })
    }
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
            ActiveState::Unspecified => out.write_all(b"unspecified")?,
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
            b"unspecified" => Ok(ActiveState::Unspecified),
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

#[derive(Debug, Clone, Copy, PartialEq, FromSqlRow, AsExpression)]
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

impl SecretType {
    /// A list of possible variants in `&'static str` form
    pub fn variants() -> [&'static str; 11] {
        [
            "unspecified",
            "basic_auth",
            "api_key",
            "env_var",
            "file",
            "ssh_key",
            "docker_registry",
            "npm_repo",
            "pypi_registry",
            "maven_repo",
            "kubernetes",
        ]
    }
}

impl FromStr for SecretType {
    type Err = std::string::ParseError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Ok(match s.to_lowercase().as_ref() {
            "unspecified" => SecretType::Unspecified,
            "basic_auth" => SecretType::BasicAuth,
            "api_key" => SecretType::ApiKey,
            "env_var" => SecretType::EnvVar,
            "file" => SecretType::File,
            "ssh_key" => SecretType::SshKey,
            "docker_registry" => SecretType::DockerRegistry,
            "npm_repo" => SecretType::NpmRepo,
            "pypi_registry" => SecretType::PypiRegistry,
            "maven_repo" => SecretType::MavenRepo,
            "kubernetes" => SecretType::Kubernetes,
            _ => SecretType::Unspecified,
        })
    }
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
            SecretType::Unspecified => out.write_all(b"unspecified")?,
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
            b"unspecified" => Ok(SecretType::Unspecified),
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

#[derive(Debug, Clone, Copy, PartialEq, FromSqlRow, AsExpression)]
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
            GitHostType::Unspecified => out.write_all(b"unspecified")?,
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
            b"unspecified" => Ok(GitHostType::Unspecified),
            b"generic" => Ok(GitHostType::Generic),
            b"bitbucket" => Ok(GitHostType::Bitbucket),
            b"github" => Ok(GitHostType::Github),
            _ => Err("Unrecognized GitHostType variant".into()),
        }
    }
}
