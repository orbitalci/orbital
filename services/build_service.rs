use orbital_headers::build_meta::{
    server::BuildService, BuildLogResponse, BuildMetadata, BuildSummaryRequest,
    BuildSummaryResponse, BuildTarget,
};
use tonic::{Request, Response, Status};

use tokio::sync::mpsc;

use crate::OrbitalServiceError;
use agent_runtime::docker;
use config_parser::yaml as parser;
use git_meta::clone;

use super::OrbitalApi;

use log::debug;

/// Implementation of protobuf derived `BuildService` trait
#[tonic::async_trait]
impl BuildService for OrbitalApi {
    /// Start a build in a container. (Stay focused.)
    async fn build_start(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        let response = Response::new(BuildMetadata::default());

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

        debug!("Cloning code into temp directory");
        let git_repo =
            clone::clone_temp_dir(&unwrapped_request.remote_uri).expect("Unable to clone repo");

        debug!("Loading orb.yml from path {:?}", &git_repo.as_path());
        let config =
            parser::load_orb_yaml(format!("{}/{}", &git_repo.as_path().display(), "orb.yml"))
                .expect("Unable to load orb.yml");

        debug!("Pulling container: {:?}", config.image.clone());
        match docker::container_pull(config.image.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(_) => {
                return Err(Status::new(
                    tonic::Code::Aborted,
                    &format!("Could not pull image {}", &config.image),
                ))
            }
        };

        // TODO: Inject the dynamic build env vars here
        let envs_vec = crate::parse_envs_input(&None);
        let vols_vec = crate::parse_volumes_input(&None);

        // Create a new container
        debug!("Creating container");
        let default_command_w_timeout = vec!["sleep", "1h"];
        let container_id = match docker::container_create(
            config.image.as_str(),
            default_command_w_timeout,
            envs_vec,
            vols_vec,
        ) {
            Ok(container_id) => container_id,
            Err(_) => {
                return Err(OrbitalServiceError::new(&format!(
                    "Could not create image {}",
                    &config.image
                ))
                .into())
            }
        };

        // Start a docker container

        match docker::container_start(&container_id) {
            Ok(container_id) => container_id,
            Err(_) => {
                return Err(OrbitalServiceError::new(&format!(
                    "Could not start image {}",
                    &config.image
                ))
                .into())
            }
        }

        // TODO: Make sure tests try to exec w/o starting the container
        // Exec into the new container
        debug!("Sending commands into container");
        for command in config.command.iter() {
            // Build the exec string
            let wrapped_command = format!("{} | tee -a /proc/1/fd/1", &command);

            let container_command = vec!["/bin/sh", "-c", wrapped_command.as_ref()];

            match docker::container_exec(container_id.as_ref(), container_command.clone()) {
                Ok(output) => {
                    debug!("Command: {:?}", &command);
                    debug!("Output: {:?}", &output);
                    output
                }
                Err(_) => {
                    return Err(OrbitalServiceError::new(&format!(
                        "Could not create image {}",
                        &config.image
                    ))
                    .into())
                }
            }
        }

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
