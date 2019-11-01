use log::debug;
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct OrbitalConfig {
    pub image: String,
    pub command: Vec<String>,
}

pub fn load_orb_yaml(path: String) -> Result<OrbitalConfig, Box<dyn std::error::Error>> {
    let f = std::fs::File::open(path)?;
    let parsed: OrbitalConfig = serde_yaml::from_reader(&f)?;

    debug!("{:?}", parsed);

    Ok(parsed)
}
