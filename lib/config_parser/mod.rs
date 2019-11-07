use serde::{Deserialize, Serialize};

/// Yaml config parser for Orbital
pub mod yaml;

/// Orbital config struct for `orb.yml`
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct OrbitalConfig {
    /// Docker image string
    pub image: String,
    /// List of commands to be executed
    pub command: Vec<String>,
}
