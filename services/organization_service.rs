use orbital_headers::organization::{
    server::OrganizationService, OrgAddRequest, OrgEntry, OrgGetRequest, OrgListResponse,
    OrgRemoveRequest, OrgUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

use orbital_database::postgres;

/// Implementation of protobuf derived `OrganizationService` trait
#[tonic::async_trait]
impl OrganizationService for OrbitalApi {
    async fn org_add(&self, request: Request<OrgAddRequest>) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();

        let pg_conn = postgres::client::establish_connection();

        let db_result = postgres::client::new_org(&pg_conn, &unwrapped_request.name)
            .expect("There was a problem adding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_get(&self, request: Request<OrgGetRequest>) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();

        let pg_conn = postgres::client::establish_connection();

        let db_result = postgres::client::get_org(&pg_conn, &unwrapped_request.name)
            .expect("There was a problem finding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_update(
        &self,
        request: Request<OrgUpdateRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();

        let pg_conn = postgres::client::establish_connection();

        let mut update_org = postgres::org::NewOrg::default();
        update_org.active_state = unwrapped_request.active_state.into();

        // Check if we've supplied a new name. Otherwise we should make sure update_org has the same name
        update_org.name = match &unwrapped_request.update_name.len() {
            0 => unwrapped_request.name.clone(),
            _ => unwrapped_request.update_name.clone(),
        };

        let db_result = postgres::client::update_org(&pg_conn, &unwrapped_request.name, update_org)
            .expect("There was a problem finding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_remove(
        &self,
        request: Request<OrgRemoveRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();

        let pg_conn = postgres::client::establish_connection();

        let remove_org = postgres::org::Org::default();

        let db_result = postgres::client::remove_org(&pg_conn, &unwrapped_request.name)
            .expect("There was a problem finding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_list(&self, request: Request<()>) -> Result<Response<OrgListResponse>, Status> {
        let unwrapped_request = request.into_inner();

        let pg_conn = postgres::client::establish_connection();

        let update_org = postgres::org::Org::default();

        // Convert the Vec<Org> response into the proto codegen version.
        let db_result: Vec<OrgEntry> = postgres::client::list_org(&pg_conn)
            .expect("There was a problem finding org in database")
            .into_iter()
            .map(|o| o.into())
            .collect();

        Ok(Response::new(OrgListResponse { orgs: db_result }))
    }
}
