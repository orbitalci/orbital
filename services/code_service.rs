use orbital_headers::code::{
    server::CodeService, GitProviderAddRequest, GitProviderEntry, GitProviderGetRequest,
    GitProviderListRequest, GitProviderListResponse, GitProviderRemoveRequest,
    GitProviderUpdateRequest, GitRepoAddRequest, GitRepoEntry, GitRepoGetRequest,
    GitRepoListRequest, GitRepoListResponse, GitRepoRemoveRequest, GitRepoUpdateRequest,
};

use orbital_headers::orbital_types::*;

use tonic::{Request, Response, Status};

use super::OrbitalApi;

use log::debug;

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
        request: Request<GitRepoGetRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        debug!("Got request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // Connect to database. Query for the repo
        // let git_repo = db_get_repo(org, git_provider, git_repo_name)-> GitRepoEntry

        let mut git_repo = GitRepoEntry::default();
        git_repo.org = unwrapped_request.org;
        git_repo.user = "git".into(); 
        git_repo.git_provider = unwrapped_request.git_provider;
        git_repo.name = unwrapped_request.name;
        git_repo.secret_type = SecretType::SshKey.into();
        git_repo.uri = unwrapped_request.uri;
        git_repo.auth_data =
            "secret/orbital/default_org/sshkey/github.com/level11consulting/orbitalci".into();

        debug!("Response: {:?}", &git_repo);
        Ok(Response::new(git_repo))
    }

    async fn git_repo_update(
        &self,
        _request: Request<GitRepoUpdateRequest>,
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
