//use orbital_agent::AgentRuntimeError;
//use log::debug;
use std::error::Error;
use std::fmt;
use thiserror::Error;

// TODO: Create some From<S> traits for protobuf types
use state_machine;

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

pub enum ServiceType {
    Build,
    Code,
    Notify,
    Org,
    Secret,
}
// FIXME: This is called URI, but in fact is just a host:port. Need to figure out how to let server and client use this default
/// Default URI for the Orbital service
pub const ORB_DEFAULT_URI: &str = "127.0.0.1:50051";

/// Number of seconds in an hour
pub const DEFAULT_BUILD_TIMEOUT: u64 = 60 * 60;

/// Return the uri for a given service
pub fn get_service_uri(_svc: ServiceType) -> &'static str {
    // TODO: Connect to consul and return a uri of the specified service

    ORB_DEFAULT_URI
}

/// Bare struct for implmenting gRPC service traits
#[derive(Clone, Debug, Default)]
pub struct OrbitalApi {}

#[derive(Debug, Error)]
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

impl From<Box<dyn Error>> for OrbitalServiceError {
    fn from(error: Box<dyn Error>) -> Self {
        OrbitalServiceError::new(&error.to_string())
    }
}

impl From<anyhow::Error> for OrbitalServiceError {
    fn from(error: anyhow::Error) -> Self {
        OrbitalServiceError::new(&error.to_string())
    }
}

impl From<orbital_agent::AgentRuntimeError> for OrbitalServiceError {
    fn from(error: orbital_agent::AgentRuntimeError) -> Self {
        OrbitalServiceError::new(&error.to_string())
    }
}

impl From<tonic::Status> for OrbitalServiceError {
    fn from(error: tonic::Status) -> Self {
        OrbitalServiceError::new(&error.message().to_string())
    }
}

impl From<tonic::transport::Error> for OrbitalServiceError {
    fn from(error: tonic::transport::Error) -> Self {
        OrbitalServiceError::new(format!("{}", error).as_ref())
    }
}

impl From<OrbitalServiceError> for tonic::Status {
    fn from(error: OrbitalServiceError) -> Self {
        tonic::Status::new(tonic::Code::Aborted, &error.details)
    }
}
