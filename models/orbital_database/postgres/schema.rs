// With our custom PG enums, we cannot allow diesel-cli to modify this file.
// See: diesel-rs/diesel#343
use diesel::deserialize::{self, FromSql};
use diesel::pg::Pg;
use diesel::serialize::{self, IsNull, Output, ToSql};
use std::io::Write;

use orbital_headers::orbital_types;

// Custom type handling reference:
// https://github.com/diesel-rs/diesel/blob/1.4.x/diesel_tests/tests/custom_types.rs

table! {
    use diesel::sql_types::{Integer, Text, Timestamp};
    use super::ActiveStatePGEnum;

    org {
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
