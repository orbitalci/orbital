use orbital_headers::organization::{
    server::OrganizationService, Org, OrgDeleteRequest, OrgDisableRequest, OrgEnableRequest,
    OrgRegisterRequest, PolledRepo, RegisteredRepo, RegisteredRepoUpdateStateRequest,
    RegisteredRepoUpdateUriRequest, RepoRegisterPollingExpressionRequest, RepoRegisterRequest,
    RepoUpdatePollingStateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `OrganizationService` trait
#[tonic::async_trait]
impl OrganizationService for OrbitalApi {
    async fn register_org(
        &self,
        _request: Request<OrgRegisterRequest>,
    ) -> Result<Response<Org>, Status> {
        unimplemented!()
    }

    async fn enable_org(
        &self,
        _request: Request<OrgEnableRequest>,
    ) -> Result<Response<Org>, Status> {
        unimplemented!()
    }

    async fn disable_org(
        &self,
        _request: Request<OrgDisableRequest>,
    ) -> Result<Response<Org>, Status> {
        unimplemented!()
    }

    async fn delete_org(
        &self,
        _request: Request<OrgDeleteRequest>,
    ) -> Result<Response<Org>, Status> {
        unimplemented!()
    }

    async fn register_repo(
        &self,
        _request: Request<RepoRegisterRequest>,
    ) -> Result<Response<RegisteredRepo>, Status> {
        unimplemented!()
    }

    async fn update_repo_state(
        &self,
        _request: Request<RegisteredRepoUpdateStateRequest>,
    ) -> Result<Response<RegisteredRepo>, Status> {
        unimplemented!()
    }

    async fn update_repo_uri(
        &self,
        _request: Request<RegisteredRepoUpdateUriRequest>,
    ) -> Result<Response<RegisteredRepo>, Status> {
        unimplemented!()
    }

    async fn poll_repo(
        &self,
        _request: Request<RepoRegisterPollingExpressionRequest>,
    ) -> Result<Response<PolledRepo>, Status> {
        unimplemented!()
    }

    async fn update_repo_polling_state(
        &self,
        _request: Request<RepoUpdatePollingStateRequest>,
    ) -> Result<Response<PolledRepo>, Status> {
        unimplemented!()
    }
}
