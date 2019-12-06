use hashicorp_vault as vault;

use log::debug;

pub fn add_secret(vault_host: &str, vault_token: &str, path: &str, data: &str) -> Result<(), ()> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

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

pub fn get_secret(vault_host: &str, vault_token: &str, path: &str) -> Result<String, ()> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

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

// This is a copy of add_secret for now
pub fn update_secret(
    vault_host: &str,
    vault_token: &str,
    path: &str,
    data: &str,
) -> Result<(), ()> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

    match client.set_secret(path, data) {
        Ok(_) => {
            debug!("Secret was updated");
            Ok(())
        }
        Err(_e) => {
            panic!("There was an error updating the secret");
        }
    }
}

pub fn remove_secret(vault_host: &str, vault_token: &str, path: &str) -> Result<(), ()> {
    let client = vault::Client::new(vault_host, vault_token).unwrap();

    let _ = match client.delete_secret(path) {
        Ok(secret) => {
            debug!("Found secret and deleted it");
            secret
        }
        Err(_e) => {
            panic!("There was an error deleting the secret");
        }
    };

    Ok(())
}

// TODO: The underlying vault library needs to add this function for secrets
//pub fn list_secret(vault_host: &str, vault_token: &str, path: &str) -> Result<(), ()> {
//    let host = match env::var("VAULT_ADDR") {
//        Ok(val) => val,
//        Err(_e) => {
//            debug!("VAULT_ADDR env var not set. Assuming 'http://127.0.0.1:8200'");
//            "http://127.0.0.1:8200".to_string()
//        }
//    };
//
//    let token = match env::var("VAULT_TOKEN") {
//        Ok(val) => val,
//        Err(_e) => {
//            debug!("VAULT_TOKEN env var not set. Assuming 'orbital'");
//            "orbital".to_string()
//        }
//    };
//    let client = vault::Client::new(host.as_str(), token).unwrap();
//
//    let secrets = match client.list_secret(path) {
//        Ok(secret) => {
//            debug!("Found secret and deleted it");
//            secret
//        }
//        Err(_e) => {
//            panic!("There was an error deleting the secret");
//        }
//    };
//
//    Ok(secrets)
//}
