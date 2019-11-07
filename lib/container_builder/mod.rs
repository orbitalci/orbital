/// Docker engine api wrapper
pub mod docker;
/// Default volume mount mapping for host Docker into container for Docker-in-Docker builds
pub const DOCKER_SOCKET_VOLMAP: &str = "/var/run/docker.sock:/var/run/docker.sock";
/// Default working directory for staging repo code inside container
pub const ORBITAL_CONTAINER_WORKDIR: &str = "/orbital-work";
