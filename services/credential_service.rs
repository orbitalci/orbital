use futures::future::FutureResult;
use orbital_headers::credential::{
    server::CredentialService, VcsCredCreateRequest, VcsCredDeleteRequest, VcsCredEntry,
    VcsCredListRequest, VcsCredListResponse, VcsCredUpdateRequest,
};
use tower_grpc::Response;

use super::OrbitalApi;

/// Implementation of protobuf derived `CredentialService` trait
impl CredentialService for OrbitalApi {
    type CreateVcsCredFuture = FutureResult<Response<VcsCredEntry>, tower_grpc::Status>;
    type DeleteVcsCredFuture = FutureResult<Response<VcsCredEntry>, tower_grpc::Status>;
    type UpdateVcsCredFuture = FutureResult<Response<VcsCredEntry>, tower_grpc::Status>;
    type ListVcsCredsFuture = FutureResult<Response<VcsCredListResponse>, tower_grpc::Status>;

    fn create_vcs_cred(
        &mut self,
        _request: tower_grpc::Request<VcsCredCreateRequest>,
    ) -> Self::CreateVcsCredFuture {
        unimplemented!()
    }

    fn delete_vcs_cred(
        &mut self,
        _request: tower_grpc::Request<VcsCredDeleteRequest>,
    ) -> Self::DeleteVcsCredFuture {
        unimplemented!()
    }

    fn update_vcs_cred(
        &mut self,
        _request: tower_grpc::Request<VcsCredUpdateRequest>,
    ) -> Self::UpdateVcsCredFuture {
        unimplemented!()
    }

    fn list_vcs_creds(
        &mut self,
        _request: tower_grpc::Request<VcsCredListRequest>,
    ) -> Self::ListVcsCredsFuture {
        unimplemented!()
    }
}
