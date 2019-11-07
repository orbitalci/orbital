use crate::postgres::schema::{org, ActiveState};
use chrono::{NaiveDateTime, Utc};

use orbital_headers::organization::OrgEntry;

#[derive(Insertable, Debug, PartialEq, AsChangeset)]
#[table_name = "org"]
pub struct NewOrg {
    pub name: String,
    pub created: NaiveDateTime,
    pub last_update: NaiveDateTime,
    pub active_state: ActiveState,
}

impl Default for NewOrg {
    fn default() -> Self {
        NewOrg {
            name: "".into(),
            created: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            last_update: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            active_state: ActiveState::Enabled,
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable)]
#[table_name = "org"]
pub struct Org {
    pub id: i32,
    pub name: String,
    pub created: NaiveDateTime,
    pub last_update: NaiveDateTime,
    pub active_state: ActiveState,
}

impl Default for Org {
    fn default() -> Self {
        Org {
            id: 0,
            name: "".into(),
            created: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            last_update: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            active_state: ActiveState::Enabled,
        }
    }
}

impl From<OrgEntry> for Org {
    fn from(org: OrgEntry) -> Self {
        // TODO: Need to convert timestamps from OrgEntry

        Org {
            id: org.id.into(),
            name: org.name.into(),
            created: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            last_update: NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0),
            active_state: org.active_state.into(),
        }
    }
}

impl From<Org> for OrgEntry {
    fn from(org: Org) -> Self {
        let created_timestamp = prost_types::Timestamp {
            seconds: org.created.timestamp(),
            nanos: org.created.timestamp_nanos() as i32,
        };

        let last_update_timestamp = prost_types::Timestamp {
            seconds: org.last_update.timestamp(),
            nanos: org.last_update.timestamp_nanos() as i32,
        };

        OrgEntry {
            id: org.id,
            name: org.name,
            created: Some(created_timestamp),
            last_update: Some(last_update_timestamp),
            active_state: org.active_state.into(),
        }
    }
}
