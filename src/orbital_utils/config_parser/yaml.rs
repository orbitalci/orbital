use super::OrbitalConfig;
use color_eyre::eyre::Result;
use std::path::Path;
use tracing::debug;

/// Load Orbital config file from path, parse with `serde_yaml`, return `Result<OrbitalConfig>`
pub fn load_orb_yaml(path: &Path) -> Result<OrbitalConfig> {
    let f = std::fs::File::open(path)?;
    let parsed: OrbitalConfig = serde_yaml::from_reader(&f)?;

    debug!("{:?}", parsed);

    Ok(parsed)
}

/// Load Orbital config from &str, parse with `serde_yaml`, return `Result<OrbitalConfig>`
pub fn load_orb_yaml_from_str(config: &str) -> Result<OrbitalConfig> {
    let parsed: OrbitalConfig = serde_yaml::from_str(config)?;

    debug!("{:?}", parsed);

    Ok(parsed)
}
