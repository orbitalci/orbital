// With our custom PG enums, we cannot allow diesel-cli to modify this file.
// See: diesel-rs/diesel#343
use diesel::deserialize::{self, FromSql};
use diesel::pg::Pg;
use diesel::serialize::{self, IsNull, Output, ToSql};
use std::io::Write;

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
    Unknown,
    Enabled,
    Disabled,
}

impl ToSql<ActiveStatePGEnum, Pg> for ActiveState {
    fn to_sql<W: Write>(&self, out: &mut Output<W, Pg>) -> serialize::Result {
        match *self {
            ActiveState::Unknown => out.write_all(b"unknown")?,
            ActiveState::Enabled => out.write_all(b"enabled")?,
            ActiveState::Disabled => out.write_all(b"disabled")?,
        }
        Ok(IsNull::No)
    }
}

impl FromSql<ActiveStatePGEnum, Pg> for ActiveState {
    fn from_sql(bytes: Option<&[u8]>) -> deserialize::Result<Self> {
        match not_none!(bytes) {
            b"unknown" => Ok(ActiveState::Unknown),
            b"enabled" => Ok(ActiveState::Enabled),
            b"disabled" => Ok(ActiveState::Disabled),
            _ => Err("Unrecognized ActiveState variant".into()),
        }
    }
}
