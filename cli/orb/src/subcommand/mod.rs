//#[macro_use]
extern crate structopt;
use structopt::StructOpt;

pub mod build_cmd;
pub mod cancel;
pub mod completion;
pub mod developer;
pub mod logs;
pub mod operator;
pub mod org;
pub mod poll;
pub mod repo;
pub mod secret;
pub mod summary;

use std::error::Error;
use std::fmt;

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
}

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

#[derive(Debug, StructOpt)]
#[structopt(name = "orb")]
pub struct SubcommandContext {
    #[structopt(subcommand)]
    pub subcommand: Subcommand,
    #[structopt(flatten)]
    pub global_option: GlobalOption,
}
