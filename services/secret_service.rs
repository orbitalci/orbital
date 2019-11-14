use orbital_headers::secret::{
    server::SecretService, SecretCreateRequest, SecretDeleteRequest, SecretEntry,
    SecretListRequest, SecretListResponse, SecretUpdateRequest, VcsCredCreateRequest,
    VcsCredDeleteRequest, VcsCredEntry, VcsCredListRequest, VcsCredListResponse,
    VcsCredUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `SecretService` trait
#[tonic::async_trait]
impl SecretService for OrbitalApi {
    async fn create_secret(
        &self,
        _request: Request<SecretCreateRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        unimplemented!()
    }

    async fn delete_secret(
        &self,
        _request: Request<SecretDeleteRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        unimplemented!()
    }

    async fn update_secret(
        &self,
        _request: Request<SecretUpdateRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        unimplemented!()
    }

    async fn list_secret(
        &self,
        _request: Request<SecretListRequest>,
    ) -> Result<Response<SecretListResponse>, Status> {
        unimplemented!()
    }

    async fn create_vcs_cred(
        &self,
        _request: Request<VcsCredCreateRequest>,
    ) -> Result<Response<VcsCredEntry>, Status> {
        unimplemented!()
    }

    async fn delete_vcs_cred(
        &self,
        _request: Request<VcsCredDeleteRequest>,
    ) -> Result<Response<VcsCredEntry>, Status> {
        unimplemented!()
    }

    async fn update_vcs_cred(
        &self,
        _request: Request<VcsCredUpdateRequest>,
    ) -> Result<Response<VcsCredEntry>, Status> {
        unimplemented!()
    }

    async fn list_vcs_creds(
        &self,
        _request: Request<VcsCredListRequest>,
    ) -> Result<Response<VcsCredListResponse>, Status> {
        unimplemented!()
    }
}
