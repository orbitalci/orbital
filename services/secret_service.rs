use orbital_headers::secret::{
    server::SecretService, SecretAddRequest, SecretEntry, SecretGetRequest, SecretListRequest,
    SecretListResponse, SecretRemoveRequest, SecretUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

use hashicorp_stack::vault;
use log::debug;

/// Implementation of protobuf derived `SecretService` trait
#[tonic::async_trait]
impl SecretService for OrbitalApi {
    async fn secret_add(
        &self,
        request: Request<SecretAddRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        debug!("secret add request: {:?}", &request);

        let unwrapped_request = request.into_inner();
        let vault_path = &unwrapped_request.name;

        // TODO: Handle errors
        let _ = vault::vault_add_secret(
            &vault_path,
            &String::from_utf8_lossy(&unwrapped_request.data),
        );

        let secret_result = SecretEntry {
            //org: "default_org".into(),
            name: vault_path.to_string(),
            secret_type: unwrapped_request.secret_type,
            ..Default::default()
        };

        Ok(Response::new(secret_result))
    }

    async fn secret_get(
        &self,
        request: Request<SecretGetRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        debug!("secret get request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // TODO: Handle errors
        let secret = vault::vault_get_secret(&unwrapped_request.name);

        let secret_result = SecretEntry {
            //org: "default_org".into(),
            name: unwrapped_request.name,
            secret_type: unwrapped_request.secret_type,
            data: secret.expect("Error unwrapping secret from Vault").into(),
            ..Default::default()
        };

        Ok(Response::new(secret_result))
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
