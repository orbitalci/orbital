use crate::orbital_headers::secret::{
    secret_service_server::SecretService, SecretAddRequest, SecretEntry, SecretGetRequest,
    SecretListRequest, SecretListResponse, SecretRemoveRequest, SecretUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

use crate::orbital_database::postgres;
use crate::orbital_database::postgres::schema::SecretType;
use crate::orbital_utils::orbital_agent::vault;
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

        // TODO: Agent runtime needs to provide this
        // If no org provided. Try to reference an org
        // If there are no orgs, throw an error
        // If there is only one org, choose it
        // If there are more than one, we need to choose a default or to throw an error

        let vault_path = vault::orb_vault_path(
            &unwrapped_request.org,
            &unwrapped_request.name,
            &crate::orbital_headers::orbital_types::SecretType::from(unwrapped_request.secret_type)
                .to_string(),
        );

        debug!("Got vault path: {:?}", &vault_path);

        // TODO: Handle errors
        let _ = vault::vault_add_secret(
            &vault_path,
            &String::from_utf8_lossy(&unwrapped_request.data),
        );

        // Add Secret reference into DB

        let orb_db =
            postgres::client::OrbitalDBClient::new().set_org(Some(unwrapped_request.org.clone()));

        let db_result = orb_db
            .secret_add(
                &unwrapped_request.name,
                unwrapped_request.secret_type.into(),
            )
            .expect("There was a problem adding secret in database");

        let secret_db = db_result.0;
        let org_db = db_result.1;

        // TODO: We want the vault path available
        let secret_result = SecretEntry {
            id: secret_db.id,
            org: org_db.name,
            name: secret_db.name,
            secret_type: secret_db.secret_type.into(),
            vault_path: secret_db.vault_path,
            active_state: secret_db.active_state.into(),
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
        let orb_db =
            postgres::client::OrbitalDBClient::new().set_org(Some(unwrapped_request.org.clone()));

        let db_result = orb_db
            .secret_get(
                &unwrapped_request.name,
                unwrapped_request.secret_type.into(),
            )
            .expect("There was a problem getting secret in database");

        let secret_db = db_result.0;
        let org_db = db_result.1;

        debug!("Requesting secret from path: {:?}", &secret_db.vault_path);
        let secret = vault::vault_get_secret(&secret_db.vault_path);

        let secret_result = SecretEntry {
            id: secret_db.id,
            org: org_db.name,
            name: secret_db.name,
            secret_type: unwrapped_request.secret_type,
            data: secret.expect("Error unwrapping secret from Vault").into(),
            vault_path: secret_db.vault_path,
            active_state: secret_db.active_state.into(),
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
            &crate::orbital_headers::orbital_types::SecretType::from(unwrapped_request.secret_type)
                .to_string(),
        );

        // TODO: Handle errors
        let _secret = vault::vault_update_secret(
            vault_path,
            &String::from_utf8_lossy(&unwrapped_request.data),
        );

        // Remove Secret reference into DB

        let orb_db =
            postgres::client::OrbitalDBClient::new().set_org(Some(unwrapped_request.org.clone()));
        // FIXME: Need a cleaner way to get org id for this struct, because we're making duplicate calls to db for org.id
        let org = orb_db
            .org_get(&unwrapped_request.org)
            .expect("Unable to get org from db");

        let secret_update = postgres::secret::NewSecret {
            name: unwrapped_request.name.clone(),
            org_id: org.id,
            secret_type: postgres::schema::SecretType::from(unwrapped_request.secret_type),
            vault_path: vault_path.to_string(),
            active_state: postgres::schema::ActiveState::from(unwrapped_request.active_state),
        };

        let _db_result = orb_db
            .secret_update(&unwrapped_request.name, secret_update)
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

        let orb_db =
            postgres::client::OrbitalDBClient::new().set_org(Some(unwrapped_request.org.clone()));

        let db_result = orb_db
            .secret_remove(
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
        let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(unwrapped_request.org));

        let filters: Option<SecretType> = None;

        // Convert the Vec<Secret> response into the proto codegen version.
        let db_result: Vec<SecretEntry> = orb_db
            .secret_list(filters)
            .expect("There was a problem listing secret from database")
            .into_iter()
            .map(|(s, o)| {
                let mut secret = SecretEntry::from(s);
                secret.org = o.name;
                secret
            })
            .collect();

        Ok(Response::new(SecretListResponse {
            secret_entries: db_result,
        }))
    }
}
