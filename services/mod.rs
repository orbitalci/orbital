use log::debug;
use std::env;
use std::error::Error;
use std::fmt;

/// gRPC service for building code
pub mod build_service;
/// gRPC service for source code integration
pub mod code_service;
/// gRPC service for external data integration
pub mod notify_service;
/// gRPC service for organization level resource management
pub mod organization_service;
/// gRPC service for secrets CRUD
pub mod secret_service;

/// Bare struct for implmenting gRPC service traits
#[derive(Clone, Debug, Default)]
pub struct OrbitalApi {}

#[derive(Debug)]
pub struct OrbitalServiceError {
    details: String,
}

impl OrbitalServiceError {
    pub fn new(msg: &str) -> OrbitalServiceError {
        OrbitalServiceError {
            details: msg.to_string(),
        }
    }
}

impl fmt::Display for OrbitalServiceError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.details)
    }
}

impl Error for OrbitalServiceError {
    fn description(&self) -> &str {
        &self.details
    }

    fn source(&self) -> Option<&(dyn Error + 'static)> {
        // Generic error, underlying cause isn't tracked.
        None
    }
}

impl From<Box<dyn Error>> for OrbitalServiceError {
    fn from(error: Box<dyn Error>) -> Self {
        OrbitalServiceError::new(&error.to_string())
    }
}

impl From<tonic::Status> for OrbitalServiceError {
    fn from(error: tonic::Status) -> Self {
        OrbitalServiceError::new(&error.message().to_string())
    }
}

impl From<OrbitalServiceError> for tonic::Status {
    fn from(error: OrbitalServiceError) -> Self {
        tonic::Status::new(tonic::Code::Aborted, &error.details)
    }
}

// Below is copied from orbital_cli_subcommand crate

/// Wrapper function for `kv_csv_parser` to specifically handle env vars for `shiplift`
pub fn parse_envs_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let envs = kv_csv_parser(user_input);
    debug!("Env vars to set: {:?}", envs);
    envs
}

/// Returns a `String` of the current working directory.
pub fn get_current_workdir() -> String {
    let path = match env::current_dir() {
        Ok(d) => format!("{}", d.display()),
        Err(_) => String::from("."),
    };

    debug!("Current workdir on host: {}", &path);
    path
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
                    get_current_workdir(),
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
            let kv_vec: Vec<&str> = n.split(",").collect();
            return Some(kv_vec);
        }
        None => return None,
    }
}
