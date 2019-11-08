use orbital_headers::credential::{
    server::CredentialService, VcsCredCreateRequest, VcsCredDeleteRequest, VcsCredEntry,
    VcsCredListRequest, VcsCredListResponse, VcsCredUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `CredentialService` trait
#[tonic::async_trait]
impl CredentialService for OrbitalApi {
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
