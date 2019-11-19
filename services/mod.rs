use agent_runtime::AgentRuntimeError;
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

/// 1 hour, in seconds
pub const DEFAULT_BUILD_TIMEOUT: u64 = 60 * 60 * 24;

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

impl From<agent_runtime::AgentRuntimeError> for OrbitalServiceError {
    fn from(error: agent_runtime::AgentRuntimeError) -> Self {
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
