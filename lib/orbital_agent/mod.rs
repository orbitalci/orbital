/// Agent api for building
pub mod build_engine;

/// Vault api wrapper
pub mod vault;

use log::debug;
use std::error::Error;
use std::fmt;
use thiserror::Error;

// Leaving this here for when we can focus on non-docker workflows
//enum AgentRuntimeType {
//    Host,
//    Docker,
//}

#[derive(Debug, Error)]
pub struct AgentRuntimeError {
    details: String,
}

impl AgentRuntimeError {
    pub fn new(msg: &str) -> AgentRuntimeError {
        AgentRuntimeError {
            details: msg.to_string(),
        }
    }
}

impl fmt::Display for AgentRuntimeError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.details)
    }
}

impl From<Box<dyn Error>> for AgentRuntimeError {
    fn from(error: Box<dyn Error>) -> Self {
        AgentRuntimeError::new(&error.to_string())
    }
}

/// Generate a tracable string for Docker containers
pub fn generate_unique_build_id(org: &str, repo: &str, commit: &str, id: &str) -> String {
    // Arbitrary max lengths
    let org_print = if org.len() > 20 { &org[0..19] } else { org };
    let repo_print = if repo.len() > 20 { &repo[0..19] } else { repo };
    let commit_print = &commit[0..7];

    format!(
        "{org}_{repo}_{commit}_{id}",
        org = org_print,
        repo = repo_print,
        commit = commit_print,
        id = id
    )
}
