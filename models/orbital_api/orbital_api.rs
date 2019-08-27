pub mod build {
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