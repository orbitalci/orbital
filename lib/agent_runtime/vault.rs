use hashicorp_stack::vault;
use log::debug;
use std::env;

pub fn orb_vault_path(org: &str, name: &str, secret_type: &str) -> String {
    format!("orbital/{}/{}/{}", org, secret_type, name,).to_lowercase()
}

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

    match vault::add_secret(host.as_str(), token.as_str(), path, data) {
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

    let secret = match vault::get_secret(host.as_str(), token.as_str(), path) {
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

pub fn vault_update_secret(path: &str, data: &str) -> Result<(), ()> {
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

    match vault::update_secret(host.as_str(), token.as_str(), path, data) {
        Ok(_) => {
            debug!("Secret was updated");
            Ok(())
        }
        Err(_) => {
            panic!("There was an error updating the secret");
        }
    }
}

pub fn vault_remove_secret(path: &str) -> Result<(), ()> {
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

    match vault::remove_secret(host.as_str(), token.as_str(), path) {
        Ok(_) => {
            debug!("Secret was removed");
            Ok(())
        }
        Err(_) => {
            panic!("There was an error removing the secret");
        }
    }
}

//pub fn vault_list_secret(path: &str, data: &str) -> Result<(), ()> {
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
//
//}
