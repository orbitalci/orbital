use anyhow;
use structopt::StructOpt;
use thiserror::Error;

/// Send a remote call for starting a build
pub mod build_cmd;
/// Send a remote call for stopping a build
pub mod cancel;
/// For Orbital developers - direct access to internal libraries outside of production-workflows
pub mod developer;
/// Request logs
pub mod logs;
/// Operator-specific commands
pub mod operator;
/// Organization-level commands
pub mod org;
/// Git repo resource support
pub mod repo;
/// Secrets engine support
pub mod secret;
/// Historical data for users
pub mod summary;

use log::debug;
use std::env;
use std::error::Error;
use std::fmt;
use std::path::PathBuf;

use git2;

// TODO: I'd like to manage errors like this to keep error text together
// Getting this error: `(dyn std::error::Error + 'static)` cannot be sent between threads safely
//#[derive(Debug, Error)]
//pub enum SubcommandErrorEnum {
//    #[error("Error from git")]
//    Git(#[from] git2::Error),
//    #[error("Error from tonic status")]
//    TonicStatus(#[from] tonic::Status),
//    #[error("Error from tonic transport")]
//    TonicTransport(#[from] tonic::transport::Error),
//    #[error("Error from anyhow")]
//    Anyhow(#[from] anyhow::Error),
//    #[error("Error from Boxed")]
//    Boxed(#[from] Box<dyn Error>),
//    #[error("Error from I/O")]
//    Io(#[from] io::Error),
//
//    #[error("Failed to parse config: {0}")]
//    ConfigParseError(String),
//    #[error("Could not pull image: {0}")]
//    DockerClientPull(String),
//    #[error("Could not create container with image: {0}")]
//    DockerClientCreate(String),
//    #[error("Could not start Docker container id: {0}")]
//    DockerClientStart(String),
//    #[error("Could not exec into container id: {0}")]
//    DockerClientExec(String),
//    #[error("Could not stop container id: {0}")]
//    DockerClientStop(String),
//}

/// Internal error type used by all subcommand handlers. Implements `Error` trait.
#[derive(Debug, Error)]
pub struct SubcommandError {
    details: String,
}

impl SubcommandError {
    pub fn new(msg: &str) -> SubcommandError {
        SubcommandError {
            details: msg.to_string(),
        }
    }
}

impl fmt::Display for SubcommandError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.details)
    }
}

impl From<Box<dyn Error>> for SubcommandError {
    fn from(error: Box<dyn Error>) -> Self {
        SubcommandError::new(&error.to_string())
    }
}

impl From<anyhow::Error> for SubcommandError {
    fn from(error: anyhow::Error) -> Self {
        SubcommandError::new(&error.to_string())
    }
}

impl From<tonic::Status> for SubcommandError {
    fn from(error: tonic::Status) -> Self {
        SubcommandError::new(&error.message().to_string())
    }
}

impl From<tonic::transport::Error> for SubcommandError {
    fn from(error: tonic::transport::Error) -> Self {
        SubcommandError::new(format!("{}", error).as_ref())
    }
}

impl From<git2::Error> for SubcommandError {
    fn from(error: git2::Error) -> Self {
        SubcommandError::new(error.message().to_string().as_ref())
    }
}

impl From<std::io::Error> for SubcommandError {
    fn from(error: std::io::Error) -> Self {
        SubcommandError::new(error.description().to_string().as_ref())
    }
}

/// Returns a `Path` of the current working directory.
pub fn get_current_workdir() -> PathBuf {
    let pathbuf = match env::current_dir() {
        Ok(p) => p,
        Err(_e) => panic!("Could not get current working directory"),
    };

    debug!("Current workdir on host: {:?}", &pathbuf);
    pathbuf
}

/// Wrapper function for `kv_csv_parser` to specifically handle env vars for `shiplift`
pub fn parse_envs_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let envs = kv_csv_parser(user_input);
    debug!("Env vars to set: {:?}", envs);
    envs
}

/// Wrapper function for `kv_csv_parser` to specifically handle volume mounts for `shiplift`
/// Automatically add in the docker socket as defined by `agent_runtime::DOCKER_SOCKET_VOLMAP`. If we don't pass in any other volumes
///
/// For now, also assume passing in the current working directory as well
pub fn parse_volumes_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let vols = match kv_csv_parser(user_input) {
        Some(v) => {
            let mut new_vec: Vec<&str> = Vec::new();
            new_vec.push(agent_runtime::DOCKER_SOCKET_VOLMAP);
            new_vec.extend(v.clone());
            Some(new_vec)
        }
        None => {
            let mut new_vec: Vec<&str> = Vec::new();
            new_vec.push(agent_runtime::DOCKER_SOCKET_VOLMAP);

            // There's got to be a better way to handle this...
            // https://stackoverflow.com/a/30527289/1672638
            new_vec.push(Box::leak(
                format!(
                    "{}:{}",
                    get_current_workdir().display(),
                    agent_runtime::ORBITAL_CONTAINER_WORKDIR,
                )
                .into_boxed_str(),
            ));
            Some(new_vec)
        }
    };
    debug!("Volumes to mount: {:?}", &vols);
    vols
}

/// Returns an `Option<Vec<&str>>` after parsing a comma-separated string from the cli
pub fn kv_csv_parser(kv_str: &Option<String>) -> Option<Vec<&str>> {
    debug!("Parsing Option<String> input: {:?}", &kv_str);
    match kv_str {
        Some(n) => {
            let kv_vec: Vec<&str> = n.split(',').collect();
            Some(kv_vec)
        }
        None => None,
    }
}

/// Top-level subcommands for `orb`
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Subcommand {
    /// Send build signal
    Build(build_cmd::SubcommandOption),
    /// Send cancel signal
    Cancel(cancel::SubcommandOption),
    /// Get logs
    Logs(logs::SubcommandOption),
    /// Actions for Organizations
    Org(org::SubcommandOption),
    /// Actions for Repos
    Repo(repo::SubcommandOption),
    /// Do things with secrets for builds
    Secret(secret::SubcommandOption),
    /// Get summary of a repo
    Summary(summary::SubcommandOption),
    /// Administration and service settings
    #[structopt(alias = "ops")]
    Operator(operator::OperatorType),
    /// Developer level commands and settings
    #[structopt(alias = "dev")]
    Developer(developer::DeveloperType),
}

/// Global command line flags that get passed down to the final subcommand handler
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct GlobalOption {
    /// Verbose mode. Display extra debug logging
    #[structopt(long)]
    pub debug: bool,
    /// Dry-run mode. No changes will be made
    #[structopt(long)]
    pub check: bool,
}

/// Represents a single-parsed command line invocation from user
#[derive(Debug, StructOpt)]
#[structopt(name = "orb")]
pub struct SubcommandContext {
    #[structopt(subcommand)]
    pub subcommand: Subcommand,
    #[structopt(flatten)]
    pub global_option: GlobalOption,
}
