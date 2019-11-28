use hashicorp_vault as vault;

use log::debug;
use std::env;

pub fn vault_add_secret(path: &str, data: &str) -> Result<(), ()> {
    let host = match env::var("VAULT_ADDR") {
        Ok(val) => val,
        Err(_e) => {
            debug!("VAULT_ADDR env var not set. Assuming 'http://127.0.0.1:8200'");
            "http://127.0.0.1:8200".to_string()
        }
    };

    let token = match env::var("VAULT_TOKEN") {
        Ok(val) => val,
        Err(_e) => {
            debug!("VAULT_TOKEN env var not set. Assuming 'orbital'");
            "orbital".to_string()
        }
    };

    let client = vault::Client::new(host.as_str(), token).unwrap();

    match client.set_secret(path, data) {
        Ok(_) => {
            debug!("Secret was set");
            Ok(())
        }
        Err(_) => {
            panic!("There was an error setting the secret");
        }
    }
}

pub fn vault_get_secret(path: &str) -> Result<String, ()> {
    let host = match env::var("VAULT_ADDR") {
        Ok(val) => val,
        Err(_e) => {
            debug!("VAULT_ADDR env var not set. Assuming 'http://127.0.0.1:8200'");
            "http://127.0.0.1:8200".to_string()
        }
    };

    let token = match env::var("VAULT_TOKEN") {
        Ok(val) => val,
        Err(_e) => {
            debug!("VAULT_TOKEN env var not set. Assuming 'orbital'");
            "orbital".to_string()
        }
    };
    let client = vault::Client::new(host.as_str(), token).unwrap();

    let secret = match client.get_secret(path) {
        Ok(secret) => {
            debug!("Found secret");
            secret
        }
        Err(_e) => {
            panic!("There was an error getting the secret");
        }
    };

    Ok(secret)
}
