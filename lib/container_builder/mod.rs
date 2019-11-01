pub mod docker;
pub const DOCKER_SOCKET_VOLMAP: &str = "/var/run/docker.sock:/var/run/docker.sock";
pub const ORBITAL_CONTAINER_WORKDIR: &str = "/orbital-work";

// Perhaps we're going to want a trait here for container builders
