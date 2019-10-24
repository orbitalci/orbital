pub mod build_service;
pub mod credential_service;
pub mod integration_service;
pub mod organization_service;

/// Bare struct for implmenting gRPC service traits
#[derive(Clone, Debug)]
pub struct OrbitalApi;