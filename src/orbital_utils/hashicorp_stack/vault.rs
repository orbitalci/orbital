use color_eyre::eyre::{eyre, Result};
use hashicorp_vault as vault;

use tracing::debug;

pub fn add_secret(vault_host: &str, vault_token: &str, path: &str, data: &str) -> Result<()> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

    match client.set_secret(path, data) {
        Ok(_) => {
            debug!("Secret was set");
            Ok(())
        }
        Err(_) => Err(eyre!("There was an error setting the secret")),
    }
}

pub fn get_secret(vault_host: &str, vault_token: &str, path: &str) -> Result<String> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

    match client.get_secret(path) {
        Ok(secret) => {
            debug!("Found secret");
            Ok(secret)
        }
        Err(_e) => Err(eyre!("There was an error getting the secret")),
    }
}

// This is a copy of add_secret for now
pub fn update_secret(vault_host: &str, vault_token: &str, path: &str, data: &str) -> Result<()> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

    match client.set_secret(path, data) {
        Ok(_) => {
            debug!("Secret was updated");
            Ok(())
        }
        Err(_e) => Err(eyre!("There was an error updating the secret")),
    }
}

pub fn remove_secret(vault_host: &str, vault_token: &str, path: &str) -> Result<()> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

    match client.delete_secret(path) {
        Ok(_secret) => {
            debug!("Found secret and deleted it");
            Ok(())
        }
        Err(_e) => Err(eyre!("There was an error deleting the secret")),
    }
}
