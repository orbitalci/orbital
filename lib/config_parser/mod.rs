use serde::{Deserialize, Serialize};
use strum_macros::{Display, EnumIter};

/// Yaml config parser for Orbital
pub mod yaml;

/*
TODO:
* Different kind of builder: docker as default, but directly on host should be offered
* commands should go inside stages
* notify
* global env vars
* stage dependent env vars
* secrets
* branches
* branch conditional stage
*/

/*
potential orb.yml
---
image: docker_image:tag
env:
secrets:
build:
  env:
  command:
    - name
      branch:
      directory:
    - notify:
        branch:
        slack:
          name:
          message:
    - notify:
        branch:
        webhook:
          name:
          message:

*/

/*
current orb.yml
---
image: docker_image:tag
command:
  - echo hello world
*/

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum OrbitalBuildMode {
    Docker,
    Host,
}

impl Default for OrbitalBuildMode {
    fn default() -> Self {
        OrbitalBuildMode::Docker
    }
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct OrbitalBranches {
    pub include: Option<Vec<String>>,
    pub exclude: Option<Vec<String>>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct OrbitalStage {
    pub name: Option<String>,
    pub command: Vec<String>,
    pub branches: Option<OrbitalBranches>,
    pub env: Option<Vec<String>>,
    // FIXME: Stage timeout. This should get parsed into a duration
    pub timeout: Option<u32>,
    pub secrets: Option<String>,
}

/// Orbital config struct for `orb.yml`
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
//#[serde(rename_all = "kebabcase")]
pub struct OrbitalConfig {
    #[serde(default)]
    pub runtime_mode: OrbitalBuildMode,
    /// Docker image string
    pub image: String,
    /// List of stages to be executed
    pub stages: Vec<OrbitalStage>,
    #[serde(default)]
    pub branches: Option<OrbitalBranches>,
    pub env: Option<Vec<String>>,
    // FIXME: Global timeout. This should get parsed into a duration
    pub timeout: Option<u32>,
    pub secrets: Option<String>,
}
