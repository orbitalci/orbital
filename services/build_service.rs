use orbital_headers::build_metadata::{
    server::BuildService, BuildDeleteRequest, BuildLogRequest, BuildLogResponse, BuildStartRequest,
    BuildStopRequest, BuildSummary,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `BuildService` trait
#[tonic::async_trait]
impl BuildService for OrbitalApi {
    /// Start a build
    async fn start_build(
        &self,
        _request: Request<BuildStartRequest>,
    ) -> Result<Response<BuildSummary>, Status> {
        let response = Response::new(BuildSummary::default());

        println!("DEBUG: {:?}", response);

        Ok(response)
    }

    async fn stop_build(
        &self,
        _request: Request<BuildStopRequest>,
    ) -> Result<Response<BuildSummary>, Status> {
        unimplemented!();
    }

    async fn get_build_logs(
        &self,
        _request: Request<BuildLogRequest>,
    ) -> Result<Response<BuildLogResponse>, Status> {
        unimplemented!();
    }

    async fn delete_build(
        &self,
        _request: Request<BuildDeleteRequest>,
    ) -> Result<Response<BuildSummary>, Status> {
        unimplemented!();
    }
}
