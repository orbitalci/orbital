// TODO: Rust's `build` module is special, and we can't override it. Need to reconcile the difference in the protos.
pub mod builder {
    include!(concat!(env!("OUT_DIR"), "/build.rs"));
}

pub mod credential {
    include!(concat!(env!("OUT_DIR"), "/credential.rs"));
}

pub mod integration {
    include!(concat!(env!("OUT_DIR"), "/integration.rs"));
}

pub mod organization {
    include!(concat!(env!("OUT_DIR"), "/organization.rs"));
}

pub mod state {
    include!(concat!(env!("OUT_DIR"), "/state.rs"));
}
