use orbital_headers::secret::{
    secret_service_server::SecretService, SecretAddRequest, SecretEntry, SecretGetRequest,
    SecretListRequest, SecretListResponse, SecretRemoveRequest, SecretUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

use agent_runtime::vault;
use log::debug;
use orbital_database::postgres;
use orbital_database::postgres::schema::SecretType;

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

        let vault_path = vault::orb_vault_path(
            &unwrapped_request.org,
            &unwrapped_request.name,
            &orbital_headers::orbital_types::SecretType::from(unwrapped_request.secret_type)
                .to_string(),
        );

        println!("Got vault path: {:?}", &vault_path);

        // TODO: Handle errors
        let _ = vault::vault_add_secret(
            &vault_path,
            &String::from_utf8_lossy(&unwrapped_request.data),
        );

        // Add Secret reference into DB

        let pg_conn = postgres::client::establish_connection();

        let db_result = postgres::client::secret_add(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
            unwrapped_request.secret_type.into(),
        )
        .expect("There was a problem adding secret in database");

        // TODO: We want the vault path available
        let secret_result = SecretEntry {
            id: db_result.id,
            org: unwrapped_request.org.into(),
            name: unwrapped_request.name.into(),
            secret_type: unwrapped_request.secret_type,
            vault_path: db_result.vault_path,
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

        // Talk to DB to get the secret path
        let pg_conn = postgres::client::establish_connection();

        let db_result = postgres::client::secret_get(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
            unwrapped_request.secret_type.into(),
        )
        .expect("There was a problem getting secret in database");

        debug!("Requesting secret from path: {:?}", &db_result.vault_path);
        let secret = vault::vault_get_secret(&db_result.vault_path);

        let secret_result = SecretEntry {
            id: db_result.id,
            org: unwrapped_request.org,
            name: unwrapped_request.name,
            secret_type: unwrapped_request.secret_type,
            data: secret.expect("Error unwrapping secret from Vault").into(),
            vault_path: db_result.vault_path,
            ..Default::default()
        };

        Ok(Response::new(secret_result))
    }

    async fn secret_update(
        &self,
        request: Request<SecretUpdateRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        debug!("secret update request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        let vault_path = &vault::orb_vault_path(
            &unwrapped_request.org,
            &unwrapped_request.name,
            &orbital_headers::orbital_types::SecretType::from(unwrapped_request.secret_type)
                .to_string(),
        );

        // TODO: Handle errors
        let _secret = vault::vault_update_secret(
            &vault_path,
            &String::from_utf8_lossy(&unwrapped_request.data),
        );

        // Remove Secret reference into DB

        let pg_conn = postgres::client::establish_connection();
        // FIXME: Need a cleaner way to get org id for this struct, because we're making duplicate calls to db for org.id
        let org = postgres::client::org_get(&pg_conn, &unwrapped_request.org)
            .expect("Unable to get org from db");

        let secret_update = postgres::secret::NewSecret {
            name: unwrapped_request.name.clone(),
            org_id: org.id,
            secret_type: postgres::schema::SecretType::from(unwrapped_request.secret_type.clone()),
            vault_path: vault_path.to_string(),
            active_state: postgres::schema::ActiveState::from(
                unwrapped_request.active_state.clone(),
            ),
        };

        let _db_result = postgres::client::secret_update(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
            secret_update,
        )
        .expect("There was a problem updating secret in database");

        let secret_result = SecretEntry {
            org: unwrapped_request.org,
            name: unwrapped_request.name,
            secret_type: unwrapped_request.secret_type,
            ..Default::default()
        };

        Ok(Response::new(secret_result))
    }

    async fn secret_remove(
        &self,
        request: Request<SecretRemoveRequest>,
    ) -> Result<Response<SecretEntry>, Status> {
        debug!("secret remove request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // Remove Secret reference into DB

        let pg_conn = postgres::client::establish_connection();

        let db_result = postgres::client::secret_remove(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
            unwrapped_request.secret_type.into(),
        )
        .expect("There was a problem removing secret in database");

        // TODO: Handle errors
        debug!(
            "Trying to delete secret from vault: {:?}",
            &db_result.vault_path
        );
        let _secret = vault::vault_remove_secret(&db_result.vault_path);

        let secret_result = SecretEntry {
            org: unwrapped_request.org,
            name: unwrapped_request.name,
            secret_type: unwrapped_request.secret_type,
            ..Default::default()
        };

        Ok(Response::new(secret_result))
    }

    async fn secret_list(
        &self,
        request: Request<SecretListRequest>,
    ) -> Result<Response<SecretListResponse>, Status> {
        let unwrapped_request = request.into_inner();
        let pg_conn = postgres::client::establish_connection();

        let filters: Option<SecretType> = None;

        // Convert the Vec<Secret> response into the proto codegen version.
        let db_result: Vec<SecretEntry> =
            postgres::client::secret_list(&pg_conn, &unwrapped_request.org, filters)
                .expect("There was a problem listing secret from database")
                .into_iter()
                .map(|o| o.into())
                .collect();

        Ok(Response::new(SecretListResponse {
            secret_entries: db_result,
        }))
    }
}
