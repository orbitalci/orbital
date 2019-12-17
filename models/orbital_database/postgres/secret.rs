use crate::postgres::schema::{secret, ActiveState, SecretType};

use orbital_headers::secret::SecretEntry;

#[derive(Insertable, Debug, PartialEq, AsChangeset)]
#[table_name = "secret"]
pub struct NewSecret {
    pub name: String,
    pub org_id: i32,
    pub secret_type: SecretType,
    pub vault_path: String,
    pub active_state: ActiveState,
}

impl Default for NewSecret {
    fn default() -> Self {
        NewSecret {
            name: "".into(),
            org_id: 0,
            secret_type: SecretType::Unspecified,
            vault_path: "".into(),
            active_state: ActiveState::Enabled,
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable)]
#[table_name = "secret"]
pub struct Secret {
    pub id: i32,
    pub name: String,
    pub org_id: i32,
    pub secret_type: SecretType,
    pub vault_path: String,
    pub active_state: ActiveState,
}

impl Default for Secret {
    fn default() -> Self {
        Secret {
            id: 0,
            name: "".into(),
            org_id: 0,
            secret_type: SecretType::Unspecified,
            vault_path: "".into(),
            active_state: ActiveState::Enabled,
        }
    }
}

impl From<Secret> for SecretEntry {
    fn from(secret: Secret) -> Self {
        SecretEntry {
            ..Default::default()
        }
    }
}
