// Dig into a filepath and harvest info about the local branch

// Get as much info about the remote branch as well

use git2::Repository;
use url::{Url, Host};


// TODO: This should return a Result so we can use questionmark operator
pub fn git_remote_from_path(path : &str) -> String {
    let p = match Repository::open(path) {
        Ok(repo) => repo,
        Err(e) => panic!("failed to init: {}", e),
    };

    let remote_url :  String = p.find_remote("origin").unwrap().url().unwrap().chars().collect();

    // Just playing around here
    let _ = git_remote_url_parse(&remote_url);


    remote_url
}


#[derive(Debug)]
pub struct GitSshRemote {
    user : String,
    provider : String,
    account : String,
    repo : String,
}

pub fn git_remote_url_parse(remote_url : &str) -> GitSshRemote {

    // We will want to see if we can parse w/ Url, since git repos might use HTTPS
    let http_url = Url::parse(remote_url);
    // If we get Err(RelativeUrlWithoutBase) then we should pick apart the remote url
    println!("{:?}",http_url);


    // Splitting on colon first will split
    // user@provider.tld:account/repo.git
    let split_first_stage = remote_url.split(":").collect::<Vec<&str>>();

    let user_provider = split_first_stage[0].split("@").collect::<Vec<&str>>();
    let acct_repo = split_first_stage[1].split("/").collect::<Vec<&str>>();

    GitSshRemote {
        user : user_provider[0].to_string(),
        provider : user_provider[1].to_string(),
        account : acct_repo[0].to_string(),
        repo : acct_repo[1].to_string(),
    }
}
