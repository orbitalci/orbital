use futures::future;
use orbital_headers::builder::{BuildLogResponse, BuildSummary};
use tower_grpc::{Request, Response};

#[derive(Clone, Debug)]
pub struct OrbitalApi;

impl orbital_headers::builder::server::BuildService for OrbitalApi {
    type StartBuildFuture = future::FutureResult<Response<BuildSummary>, tower_grpc::Status>;
    type StopBuildFuture = future::FutureResult<Response<BuildSummary>, tower_grpc::Status>;
    type GetBuildLogsFuture = future::FutureResult<Response<BuildLogResponse>, tower_grpc::Status>;
    type DeleteBuildFuture = future::FutureResult<Response<BuildSummary>, tower_grpc::Status>;

    fn start_build(
        &mut self,
        _request: Request<orbital_headers::builder::BuildStartRequest>,
    ) -> Self::StartBuildFuture {
        let response = Response::new(BuildSummary::default());

        println!("DEBUG: {:?}", response);

        future::ok(response)
    }

    fn stop_build(
        &mut self,
        _request: Request<orbital_headers::builder::BuildStopRequest>,
    ) -> Self::StopBuildFuture {
        unimplemented!();
    }

    fn get_build_logs(
        &mut self,
        _request: Request<orbital_headers::builder::BuildLogRequest>,
    ) -> Self::GetBuildLogsFuture {
        unimplemented!();
    }

    fn delete_build(
        &mut self,
        _request: Request<orbital_headers::builder::BuildDeleteRequest>,
    ) -> Self::DeleteBuildFuture {
        unimplemented!();
    }
}
