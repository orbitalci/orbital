use serde::{Deserialize, Serialize};

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

/// Orbital config struct for `orb.yml`
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct OrbitalConfig {
    /// Docker image string
    pub image: String,
    /// List of commands to be executed
    pub command: Vec<String>,
}
