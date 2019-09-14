use tower_grpc::{Request, Response};
use futures::{future, Future, Stream};
use orbital_api::organization::{server, Org, RegisteredRepo, PolledRepo};

#[derive(Clone, Debug)]
struct OrbitalApi;

impl orbital_api::organization::server::OrganizationService for OrbitalApi {
    type RegisterOrgFuture = future::FutureResult<Response<Org>, tower_grpc::Status>;
    type EnableOrgFuture = future::FutureResult<Response<Org>, tower_grpc::Status>;
    type DisableOrgFuture = future::FutureResult<Response<Org>, tower_grpc::Status>;
    type DeleteOrgFuture = future::FutureResult<Response<Org>, tower_grpc::Status>;
    type RegisterRepoFuture = future::FutureResult<Response<RegisteredRepo>, tower_grpc::Status>;
    type UpdateRepoStateFuture = future::FutureResult<Response<RegisteredRepo>, tower_grpc::Status>;
    type UpdateRepoUriFuture = future::FutureResult<Response<RegisteredRepo>, tower_grpc::Status>;
    type PollRepoFuture = future::FutureResult<Response<PolledRepo>, tower_grpc::Status>;
    type UpdateRepoPollingStateFuture = future::FutureResult<Response<PolledRepo>, tower_grpc::Status>;

    fn register_org(&mut self, request: tower_grpc::Request<orbital_api::organization::OrgRegisterRequest>) -> Self::RegisterOrgFuture {
        unimplemented!()
    }

    fn enable_org(&mut self, request: tower_grpc::Request<orbital_api::organization::OrgEnableRequest>) -> Self::EnableOrgFuture {
        unimplemented!()
    }

    fn disable_org(&mut self, request: tower_grpc::Request<orbital_api::organization::OrgDisableRequest>) -> Self::DisableOrgFuture {
        unimplemented!()
    }

    fn delete_org(&mut self, request: tower_grpc::Request<orbital_api::organization::OrgDeleteRequest>) -> Self::DeleteOrgFuture {
        unimplemented!()
    }

    fn register_repo(&mut self, request: tower_grpc::Request<orbital_api::organization::RepoRegisterRequest>) -> Self::RegisterRepoFuture {
        unimplemented!()
    }

    fn update_repo_state(&mut self, request: tower_grpc::Request<orbital_api::organization::RegisteredRepoUpdateStateRequest>) -> Self::UpdateRepoStateFuture {
        unimplemented!()
    }

    fn update_repo_uri(&mut self, request: tower_grpc::Request<orbital_api::organization::RegisteredRepoUpdateUriRequest>) -> Self::UpdateRepoUriFuture {
        unimplemented!()
    }

    fn poll_repo(&mut self, request: tower_grpc::Request<orbital_api::organization::RepoRegisterPollingExpressionRequest>) -> Self::PollRepoFuture {
        unimplemented!()
    }

    fn update_repo_polling_state(&mut self, request: tower_grpc::Request<orbital_api::organization::RepoUpdatePollingStateRequest>) -> Self::UpdateRepoPollingStateFuture {
        unimplemented!()
    }
}