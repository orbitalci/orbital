use futures::{future, Future, Stream};
use orbital_headers::credential::{
    server, VcsCredCreateRequest, VcsCredDeleteRequest, VcsCredEntry, VcsCredListRequest,
    VcsCredListResponse, VcsCredUpdateRequest,
};
use tower_grpc::{Request, Response};

#[derive(Clone, Debug)]
struct OrbitalApi;

impl orbital_headers::credential::server::CredentialService for OrbitalApi {
    type CreateVcsCredFuture = future::FutureResult<Response<VcsCredEntry>, tower_grpc::Status>;
    type DeleteVcsCredFuture = future::FutureResult<Response<VcsCredEntry>, tower_grpc::Status>;
    type UpdateVcsCredFuture = future::FutureResult<Response<VcsCredEntry>, tower_grpc::Status>;
    type ListVcsCredsFuture =
        future::FutureResult<Response<VcsCredListResponse>, tower_grpc::Status>;

    fn create_vcs_cred(
        &mut self,
        request: tower_grpc::Request<orbital_headers::credential::VcsCredCreateRequest>,
    ) -> Self::CreateVcsCredFuture {
        unimplemented!()
    }

    fn delete_vcs_cred(
        &mut self,
        request: tower_grpc::Request<orbital_headers::credential::VcsCredDeleteRequest>,
    ) -> Self::DeleteVcsCredFuture {
        unimplemented!()
    }

    fn update_vcs_cred(
        &mut self,
        request: tower_grpc::Request<orbital_headers::credential::VcsCredUpdateRequest>,
    ) -> Self::UpdateVcsCredFuture {
        unimplemented!()
    }

    fn list_vcs_creds(
        &mut self,
        request: tower_grpc::Request<orbital_headers::credential::VcsCredListRequest>,
    ) -> Self::ListVcsCredsFuture {
        unimplemented!()
    }
}
