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

#[derive(Clone, Debug, Identifiable, Queryable, QueryableByName)]
#[table_name = "secret"]
pub struct Secret {
    pub id: i32,
    pub org_id: i32,
    pub name: String,
    pub secret_type: SecretType,
    pub vault_path: String,
    pub active_state: ActiveState,
}

impl Default for Secret {
    fn default() -> Self {
        Secret {
            id: 0,
            org_id: 0,
            name: "".into(),
            secret_type: SecretType::Unspecified,
            vault_path: "".into(),
            active_state: ActiveState::Enabled,
        }
    }
}

// FIXME: Org should be a string, but right now we only have the postgres org id
impl From<Secret> for SecretEntry {
    fn from(secret: Secret) -> Self {
        SecretEntry {
            id: secret.id,
            org: secret.org_id.to_string(),
            name: secret.name,
            secret_type: secret.secret_type.into(),
            vault_path: secret.vault_path.into(),
            active_state: secret.active_state.into(),
            ..Default::default()
        }
    }
}

// FIXME: This does not correctly set the org id
impl From<SecretEntry> for Secret {
    fn from(secret_entry: SecretEntry) -> Self {
        Secret {
            id: secret_entry.id,
            org_id: 0,
            name: secret_entry.name,
            secret_type: secret_entry.secret_type.into(),
            vault_path: secret_entry.vault_path.into(),
            active_state: secret_entry.active_state.into(),
            ..Default::default()
        }
    }
}
