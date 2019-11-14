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
    /// Start a build
    async fn build_start(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        let response = Response::new(BuildMetadata::default());

        println!("DEBUG: {:?}", response);

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
