extern crate structopt;
use structopt::StructOpt;

/// Send a remote call for starting a build
pub mod build_cmd;
/// Send a remote call for stopping a build
pub mod cancel;
/// Generate command line shell completions
pub mod completion;
/// For Orbital developers - direct access to internal libraries outside of production-workflows
pub mod developer;
/// Request logs
pub mod logs;
/// Operator-specific commands
pub mod operator;
/// Organization-level commands
pub mod org;
/// Polling support
pub mod poll;
/// Git repo resource support
pub mod repo;
/// Secrets engine support
pub mod secret;
/// Historical data for users
pub mod summary;

use std::error::Error;
use std::fmt;

/// Internal error type used by all subcommand handlers. Implements `Error` trait.
#[derive(Debug)]
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

impl Error for SubcommandError {
    fn description(&self) -> &str {
        &self.details
    }

    fn source(&self) -> Option<&(dyn Error + 'static)> {
        // Generic error, underlying cause isn't tracked.
        None
    }
}

impl From<Box<dyn Error>> for SubcommandError {
    fn from(error: Box<dyn Error>) -> Self {
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

/// Top-level subcommands for `orb`
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Subcommand {
    /// Send build signal
    Build(build_cmd::SubcommandOption),
    /// Send cancel signal
    Cancel,
    /// Get logs
    Logs,
    /// Actions for Organizations
    Org,
    /// Actions for Repos
    Repo,
    /// Actions for Polling
    Poll,
    /// Do things with secrets for builds
    Secret,
    /// Get summary of a repo
    Summary,
    /// Administration and service settings
    #[structopt(alias = "ops")]
    Operator(operator::OperatorType),
    /// Developer level commands and settings
    #[structopt(alias = "dev")]
    Developer(developer::DeveloperType),
    /// Get version string
    Version,
    /// Generate shell completions script for orb command
    Completion(completion::SubcommandOption),
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
