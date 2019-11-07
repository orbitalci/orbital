use futures::{future, future::FutureResult};
use orbital_headers::build_metadata::{
    server::BuildService, BuildDeleteRequest, BuildLogRequest, BuildLogResponse, BuildStartRequest,
    BuildStopRequest, BuildSummary,
};
use tower_grpc::{Request, Response};

use super::OrbitalApi;

/// Implementation of protobuf derived `BuildService` trait
impl BuildService for OrbitalApi {
    type StartBuildFuture = FutureResult<Response<BuildSummary>, tower_grpc::Status>;
    type StopBuildFuture = FutureResult<Response<BuildSummary>, tower_grpc::Status>;
    type GetBuildLogsFuture = FutureResult<Response<BuildLogResponse>, tower_grpc::Status>;
    type DeleteBuildFuture = FutureResult<Response<BuildSummary>, tower_grpc::Status>;

    /// Start a build
    fn start_build(&mut self, _request: Request<BuildStartRequest>) -> Self::StartBuildFuture {
        let response = Response::new(BuildSummary::default());

        println!("DEBUG: {:?}", response);

        future::ok(response)
    }

    fn stop_build(&mut self, _request: Request<BuildStopRequest>) -> Self::StopBuildFuture {
        unimplemented!();
    }

    fn get_build_logs(&mut self, _request: Request<BuildLogRequest>) -> Self::GetBuildLogsFuture {
        unimplemented!();
    }

    fn delete_build(&mut self, _request: Request<BuildDeleteRequest>) -> Self::DeleteBuildFuture {
        unimplemented!();
    }
}
