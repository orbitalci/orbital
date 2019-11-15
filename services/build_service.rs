use orbital_headers::build_meta::{
    server::BuildService, BuildLogResponse, BuildMetadata, BuildSummaryRequest,
    BuildSummaryResponse, BuildTarget,
};
use tonic::{Request, Response, Status};

use tokio::sync::mpsc;

use super::OrbitalApi;

/// Implementation of protobuf derived `BuildService` trait
#[tonic::async_trait]
impl BuildService for OrbitalApi {
    /// Start a build in a container. (Stay focused.)
    async fn build_start(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        let response = Response::new(BuildMetadata::default());

        println!("DEBUG: {:?}", response);

        // Placeholder for things we need to check for eventually when we support more things
        // Org - We want the user to select their org, and cache it locally
        // Is Repo in the system? Halt if not
        // Git provider - to select APIs - a property of the repo. This is more useful if self-hosting git
        // Determine the type of secret
        // Connect to SecretService: Get the creds for the repo
        // Connect to SecretService: Collect all env vars

        // Start a docker container

        // Git clone for provider ( uri, branch, commit )
        // Copy the orb file to a temp file on host
        // Read the orb file and parse

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
