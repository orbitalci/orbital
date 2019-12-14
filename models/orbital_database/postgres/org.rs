use crate::postgres::schema::{org, ActiveState};
use chrono::{NaiveDateTime, Utc};

#[derive(Insertable, Debug, PartialEq)]
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

//impl From<org::active_state> for Activ
