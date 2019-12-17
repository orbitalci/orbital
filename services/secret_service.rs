use orbital_headers::secret::{
    server::SecretService, SecretAddRequest, SecretEntry, SecretGetRequest, SecretListRequest,
    SecretListResponse, SecretRemoveRequest, SecretUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

use agent_runtime::vault;
use log::debug;
use orbital_database::postgres;

/// Implementation of protobuf derived `SecretService` trait
#[tonic::async_trait]
impl SecretService for OrbitalApi {
    async fn secret_add(
        &self,
        request: Request<SecretAddRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        debug!("secret add request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // TODO: Agent runtime needs to provide this
        // If no org provided. Try to reference an org
        // If there are no orgs, throw an error
        // If there is only one org, choose it
        // If there are more than one, we need to choose a default or to throw an error

        // TODO: Handle errors
        let _ = vault::vault_add_secret(
            &vault::orb_vault_path(
                &unwrapped_request.org,
                &unwrapped_request.name,
                &unwrapped_request.secret_type.to_string(),
            ),
            &String::from_utf8_lossy(&unwrapped_request.data),
        );

        // Add Secret reference into DB

        let pg_conn = postgres::client::establish_connection();

        let _db_result = postgres::client::secret_add(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
            unwrapped_request.secret_type.into(),
        )
        .expect("There was a problem adding secret in database");

        let secret_result = SecretEntry {
            org: unwrapped_request.org.into(),
            name: unwrapped_request.name.into(),
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
        let secret = vault::vault_get_secret(&vault::orb_vault_path(
            &unwrapped_request.org,
            &unwrapped_request.name,
            &unwrapped_request.secret_type.to_string(),
        ));

        let secret_result = SecretEntry {
            org: unwrapped_request.org,
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
