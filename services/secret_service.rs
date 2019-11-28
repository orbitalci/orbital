use orbital_headers::secret::{
    server::SecretService, SecretAddRequest, SecretEntry, SecretGetRequest, SecretListRequest,
    SecretListResponse, SecretRemoveRequest, SecretUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `SecretService` trait
#[tonic::async_trait]
impl SecretService for OrbitalApi {
    async fn secret_add(
        &self,
        _request: Request<SecretAddRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        unimplemented!()
    }

    async fn secret_get(
        &self,
        _request: Request<SecretGetRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        unimplemented!()
    }

    async fn secret_remove(
        &self,
        _request: Request<SecretRemoveRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        unimplemented!()
    }

    async fn secret_update(
        &self,
        _request: Request<SecretUpdateRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        unimplemented!()
    }

    async fn secret_list(
        &self,
        _request: Request<SecretListRequest>,
    ) -> Result<Response<SecretListResponse>, Status> {
        unimplemented!()
    }
}
