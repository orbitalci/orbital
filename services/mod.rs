/// gRPC service for building code
pub mod build_service;
/// gRPC service for secrets CRUD
pub mod credential_service;
/// gRPC service for external data integration
pub mod integration_service;
/// gRPC service for organization level resource management
pub mod organization_service;

/// Bare struct for implmenting gRPC service traits
#[derive(Clone, Debug, Default)]
pub struct OrbitalApi {}
