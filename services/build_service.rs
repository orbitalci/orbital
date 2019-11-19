use orbital_headers::build_meta::{
    server::BuildService, BuildLogResponse, BuildMetadata, BuildSummaryRequest,
    BuildSummaryResponse, BuildTarget,
};
use tonic::{Request, Response, Status};

use tokio::sync::mpsc;

use crate::OrbitalServiceError;
use agent_runtime::{build_engine, docker};
//use config_parser::yaml as parser;
use git_meta::{clone, GitCredentials};

use super::OrbitalApi;

use log::debug;

use std::path::Path;
use std::time::Duration;

/// Implementation of protobuf derived `BuildService` trait
#[tonic::async_trait]
impl BuildService for OrbitalApi {
    /// Start a build in a container. (Stay focused.)
    async fn build_start(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        //println!("DEBUG: {:?}", response);

        // Placeholder for things we need to check for eventually when we support more things
        // Org - We want the user to select their org, and cache it locally
        // Is Repo in the system? Halt if not
        // Git provider - to select APIs - a property of the repo. This is more useful if self-hosting git
        // Determine the type of secret
        // Connect to SecretService: Get the creds for the repo
        // create write a temporary file to hold the key, so we can pass it to the git clone

        // Connect to SecretService: Collect all env vars

        // Ocelot current takes all of the secrets from an org and throws it into the container.
        // We should probably follow suit until we know better

        // Git clone for provider ( uri, branch, commit )
        let unwrapped_request = request.into_inner();
        debug!("Received request: {:?}", unwrapped_request);

        let git_creds = GitCredentials::SshKey {
            username: "git".to_string(),
            public_key: Some(Path::new("/home/telant/.ssh/id_ed25519.pub")),
            private_key: &Path::new("/home/telant/.ssh/id_ed25519"),
            passphrase: None,
        };

        debug!("Cloning code into temp directory");
        let git_repo = build_engine::clone_repo(&unwrapped_request.remote_uri, git_creds)
            .expect("Unable to clone repo");

        debug!("Loading orb.yml from path {:?}", &git_repo.as_path());
        let config = build_engine::load_orb_config(Path::new(&format!(
            "{}/{}",
            &git_repo.as_path().display(),
            "orb.yml"
        )))
        .expect("Unable to load orb.yml");

        debug!("Pulling container: {:?}", config.image.clone());
        match build_engine::docker_container_pull(config.image.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // TODO: Inject the dynamic build env vars here
        let envs_vec = agent_runtime::parse_envs_input(&None);
        let vols_vec = agent_runtime::parse_volumes_input(&None);

        // Create a new container
        debug!("Creating container");
        let container_id = match build_engine::docker_container_create(
            config.image.as_str(),
            envs_vec,
            vols_vec,
            Duration::from_secs(crate::DEFAULT_BUILD_TIMEOUT),
        ) {
            Ok(id) => id,
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // Start a docker container
        debug!("Start container");
        match build_engine::docker_container_start(config.image.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // TODO: Make sure tests try to exec w/o starting the container
        // Exec into the new container
        debug!("Sending commands into container");
        match build_engine::docker_container_exec(config.image.as_str(), config.command) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        let response = Response::new(BuildMetadata::default());
        Ok(response)
    }

    async fn build_stop(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        unimplemented!();
    }

    type BuildLogsStream = mpsc::Receiver<Result<BuildLogResponse, Status>>;
    async fn build_logs(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<Self::BuildLogsStream>, Status> {
        unimplemented!();
    }

    async fn build_remove(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        unimplemented!();
    }

    async fn build_summary(
        &self,
        _request: Request<BuildSummaryRequest>,
    ) -> Result<Response<BuildSummaryResponse>, Status> {
        unimplemented!();
    }
}
