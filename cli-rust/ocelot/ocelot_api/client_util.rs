use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_hyper::{client, util};
use tower_util::MakeService;

use std::env;
use std::fs::File;
use std::io::prelude::*;

// TODO: All these functions should be returning Result<T, E>

pub fn get_client_uri() -> http::Uri {
    let uri: http::Uri = format!("http://192.168.12.34:10000").parse().unwrap();
    uri
}

pub fn get_repo(path : Option<String>) -> String {
    let p = path
        .clone()
        .unwrap_or(env::current_dir().unwrap().to_str().unwrap().to_string());

    println!("Path to repo: {:?}", p);
    p
}

pub fn read_file(path : Option<String>) -> String {
    let mut file = File::open(path.unwrap()).unwrap();
    let mut contents = String::new();
    file.read_to_string(&mut contents).unwrap();
    contents
}

#[derive(Debug, Default)]
pub struct VcsConfig {
    client_id : String,
    client_secret : String,
    token_url : String,
    account : String,
    provider : String,
}

pub fn parse_vcs_yaml(path: String) -> VcsConfig {
    unimplemented!()
}