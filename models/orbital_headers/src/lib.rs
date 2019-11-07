// TODO: Rust's `build` module is special, and we can't override it. Need to reconcile the difference in the protos.
pub mod build_metadata {
    include!(concat!(env!("OUT_DIR"), "/build.rs"));
}

/// Generated Rust from protobufs for credential handling
pub mod credential {
    include!(concat!(env!("OUT_DIR"), "/credential.rs"));
}

/// Generated Rust from protobufs for external service integration
pub mod integration {
    include!(concat!(env!("OUT_DIR"), "/integration.rs"));
}

/// Generated Rust from protobufs for high-level units, Organizations
pub mod organization {
    include!(concat!(env!("OUT_DIR"), "/organization.rs"));
}

/// Generated Rust from protobufs for possible job states
pub mod state {
    include!(concat!(env!("OUT_DIR"), "/state.rs"));
}
