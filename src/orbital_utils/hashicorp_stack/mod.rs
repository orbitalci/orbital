use tracing::debug;

/// Hashicorp Vault helper module
pub mod vault;

pub fn orb_vault_path(org: &str, name: &str, secret_type: &str) -> String {
    let path = format!("orbital/{}/{}/{}", org, secret_type, name,).to_lowercase();

    debug!("Vault Path: {:?}", &path);
    path
}
