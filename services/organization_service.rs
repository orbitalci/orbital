use futures::future::FutureResult;
use orbital_headers::organization::{
    server::OrganizationService, Org, OrgDeleteRequest, OrgDisableRequest, OrgEnableRequest,
    OrgRegisterRequest, PolledRepo, RegisteredRepo, RegisteredRepoUpdateStateRequest,
    RegisteredRepoUpdateUriRequest, RepoRegisterPollingExpressionRequest, RepoRegisterRequest,
    RepoUpdatePollingStateRequest,
};
use tower_grpc::Response;

use super::OrbitalApi;

/// Implementation of protobuf derived `OrganizationService` trait
impl OrganizationService for OrbitalApi {
    type RegisterOrgFuture = FutureResult<Response<Org>, tower_grpc::Status>;
    type EnableOrgFuture = FutureResult<Response<Org>, tower_grpc::Status>;
    type DisableOrgFuture = FutureResult<Response<Org>, tower_grpc::Status>;
    type DeleteOrgFuture = FutureResult<Response<Org>, tower_grpc::Status>;
    type RegisterRepoFuture = FutureResult<Response<RegisteredRepo>, tower_grpc::Status>;
    type UpdateRepoStateFuture = FutureResult<Response<RegisteredRepo>, tower_grpc::Status>;
    type UpdateRepoUriFuture = FutureResult<Response<RegisteredRepo>, tower_grpc::Status>;
    type PollRepoFuture = FutureResult<Response<PolledRepo>, tower_grpc::Status>;
    type UpdateRepoPollingStateFuture = FutureResult<Response<PolledRepo>, tower_grpc::Status>;

    fn register_org(
        &mut self,
        _request: tower_grpc::Request<OrgRegisterRequest>,
    ) -> Self::RegisterOrgFuture {
        unimplemented!()
    }

    fn enable_org(
        &mut self,
        _request: tower_grpc::Request<OrgEnableRequest>,
    ) -> Self::EnableOrgFuture {
        unimplemented!()
    }

    fn disable_org(
        &mut self,
        _request: tower_grpc::Request<OrgDisableRequest>,
    ) -> Self::DisableOrgFuture {
        unimplemented!()
    }

    fn delete_org(
        &mut self,
        _request: tower_grpc::Request<OrgDeleteRequest>,
    ) -> Self::DeleteOrgFuture {
        unimplemented!()
    }

    fn register_repo(
        &mut self,
        _request: tower_grpc::Request<RepoRegisterRequest>,
    ) -> Self::RegisterRepoFuture {
        unimplemented!()
    }

    fn update_repo_state(
        &mut self,
        _request: tower_grpc::Request<RegisteredRepoUpdateStateRequest>,
    ) -> Self::UpdateRepoStateFuture {
        unimplemented!()
    }

    fn update_repo_uri(
        &mut self,
        _request: tower_grpc::Request<RegisteredRepoUpdateUriRequest>,
    ) -> Self::UpdateRepoUriFuture {
        unimplemented!()
    }

    fn poll_repo(
        &mut self,
        _request: tower_grpc::Request<RepoRegisterPollingExpressionRequest>,
    ) -> Self::PollRepoFuture {
        unimplemented!()
    }

    fn update_repo_polling_state(
        &mut self,
        _request: tower_grpc::Request<RepoUpdatePollingStateRequest>,
    ) -> Self::UpdateRepoPollingStateFuture {
        unimplemented!()
    }
}
