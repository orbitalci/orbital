/// Generated Rust from protobufs for build metadata and build services
pub mod build_metadata {
    tonic::include_proto!("build_metadata");
}

/// Generated Rust from protobufs for credential handling
/// Proto compilation issue is tracked by issue https://github.com/level11consulting/orbitalci/issues/229
pub mod credential {
    tonic::include_proto!("credential");
    tonic::include_proto!("credential_service");
}

/// Generated Rust from protobufs for external service integration
pub mod integration {
    tonic::include_proto!("integration");
}

/// Generated Rust from protobufs for high-level units, Organizations
pub mod organization {
    tonic::include_proto!("organization");
}

/// Generated Rust from protobufs for possible job states
pub mod state {
    tonic::include_proto!("state");
}
