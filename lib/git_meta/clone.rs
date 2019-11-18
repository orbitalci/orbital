use git2::{build::RepoBuilder, Cred, FetchOptions, RemoteCallbacks, Repository};
use log::debug;
use mktemp::Temp;
use std::error::Error;
use std::path::Path;

/// Create a temporary directory with mktemp, clone given uri into it.
/// Return mktemp directory, which will delete when out of scope
pub fn clone_temp_dir(uri: &str) -> Result<Temp, Box<dyn Error>> {
    let mut temp_dir = Temp::new_dir()?;

    debug!("Temp dir path: {:?}", &temp_dir.as_path());

    let mut builder = RepoBuilder::new();
    let mut callbacks = RemoteCallbacks::new();
    let mut fetch_options = FetchOptions::new();

    callbacks.credentials(|_, _, _| {
        let credentials = Cred::ssh_key(
            "git",
            Some(Path::new("/home/telant/.ssh/id_ed25519.pub")),
            Path::new("/home/telant/.ssh/id_ed25519"),
            None,
        )
        .expect("Could not create credentials object");

        Ok(credentials)
    });

    fetch_options.remote_callbacks(callbacks);

    builder.fetch_options(fetch_options);

    //let _repo = match Repository::clone(uri, &temp_dir.as_path()) {
    let _repo = match builder.clone(uri, &temp_dir.as_path()) {
        Ok(repo) => repo,
        Err(e) => panic!("failed to clone: {}", e),
    };

    Ok(temp_dir)
}
