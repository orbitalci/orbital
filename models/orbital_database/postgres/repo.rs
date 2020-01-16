use crate::postgres::schema::{repo, ActiveState, GitHostType};

use orbital_headers::secret::SecretEntry;

#[derive(Insertable, Debug, PartialEq, AsChangeset)]
#[table_name = "repo"]
pub struct NewRepo {
    pub org_id: i32,
    pub name: String,
    pub uri: String,
    pub git_host_type: GitHostType,
    pub secret_id: Option<i32>,
    pub build_active_state: ActiveState,
    pub notify_active_state: ActiveState,
    pub next_build_index: i32,
}

impl Default for NewRepo {
    fn default() -> Self {
        NewRepo {
            org_id: 0,
            name: "".into(),
            uri: "".into(),
            git_host_type: GitHostType::Generic,
            secret_id: None,
            build_active_state: ActiveState::Enabled,
            notify_active_state: ActiveState::Enabled,
            next_build_index: 0,
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable, QueryableByName)]
#[table_name = "repo"]
pub struct Repo {
    pub id: i32,
    pub org_id: i32,
    pub name: String,
    pub uri: String,
    pub git_host_type: GitHostType,
    pub secret_id: Option<i32>,
    pub build_active_state: ActiveState,
    pub notify_active_state: ActiveState,
    pub next_build_index: i32,
}

impl Default for Repo {
    fn default() -> Self {
        Repo {
            id: 0,
            org_id: 0,
            name: "".into(),
            uri: "".into(),
            git_host_type: GitHostType::Generic,
            secret_id: None,
            build_active_state: ActiveState::Enabled,
            notify_active_state: ActiveState::Enabled,
            next_build_index: 0,
        }
    }
}
