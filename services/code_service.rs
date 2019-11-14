use orbital_headers::code::{
    server::CodeService, GitProviderAddRequest, GitProviderEntry, GitProviderGetRequest,
    GitProviderListRequest, GitProviderListResponse, GitProviderRemoveRequest,
    GitProviderUpdateRequest, GitRepoAddRequest, GitRepoEntry, GitRepoGetRequest,
    GitRepoListRequest, GitRepoListResponse, GitRepoRemoveRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `CodeService` trait
#[tonic::async_trait]
impl CodeService for OrbitalApi {
    async fn git_provider_add(
        &self,
        _request: Request<GitProviderAddRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_get(
        &self,
        _request: Request<GitProviderGetRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_update(
        &self,
        _request: Request<GitProviderUpdateRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_remove(
        &self,
        _request: Request<GitProviderRemoveRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_list(
        &self,
        _request: Request<GitProviderListRequest>,
    ) -> Result<Response<GitProviderListResponse>, Status> {
        unimplemented!()
    }

    async fn git_repo_add(
        &self,
        _request: Request<GitRepoAddRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        unimplemented!()
    }

    async fn git_repo_get(
        &self,
        _request: Request<GitRepoGetRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        unimplemented!()
    }

    async fn git_repo_update(
        &self,
        _request: Request<GitProviderUpdateRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        unimplemented!()
    }

    async fn git_repo_remove(
        &self,
        _request: Request<GitRepoRemoveRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        unimplemented!()
    }

    async fn git_repo_list(
        &self,
        _request: Request<GitRepoListRequest>,
    ) -> Result<Response<GitRepoListResponse>, Status> {
        unimplemented!()
    }
}
