use super::OrbitalConfig;
use log::debug;
use std::path::Path;
use anyhow::Result;

/// Load Orbital config file from path, parse with `serde_yaml`, return `Result<OrbitalConfig>`
pub fn load_orb_yaml(path: &Path) -> Result<OrbitalConfig> {
    let f = std::fs::File::open(path)?;
    let parsed: OrbitalConfig = serde_yaml::from_reader(&f)?;

    debug!("{:?}", parsed);

    Ok(parsed)
}
