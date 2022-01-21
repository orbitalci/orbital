use crate::orbital_headers::organization::{
    organization_service_server::OrganizationService, OrgAddRequest, OrgEntry, OrgGetRequest,
    OrgListResponse, OrgRemoveRequest, OrgUpdateRequest,
};
use tracing::info;
use tonic::{Request, Response, Status};

use super::OrbitalApi;

use crate::orbital_database::postgres;

/// Implementation of protobuf derived `OrganizationService` trait
#[tonic::async_trait]
impl OrganizationService for OrbitalApi {
    async fn org_add(&self, request: Request<OrgAddRequest>) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();
        info!("org add request: {:?}", &unwrapped_request.name);

        let orb_db = postgres::client::OrbitalDBClient::new();

        let db_result = orb_db
            .org_add(&unwrapped_request.name)
            .expect("There was a problem adding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_get(&self, request: Request<OrgGetRequest>) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();
        info!("org get request: {:?}", &unwrapped_request.name);

        let orb_db = postgres::client::OrbitalDBClient::new();

        let db_result = orb_db
            .org_get(&unwrapped_request.name)
            .expect("There was a problem finding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_update(
        &self,
        request: Request<OrgUpdateRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();
        info!("org update request: {:?}", &unwrapped_request.name);

        let orb_db = postgres::client::OrbitalDBClient::new();

        let update_org = postgres::org::NewOrg {
            active_state: unwrapped_request.active_state.into(),
            name: match &unwrapped_request.update_name.len() {
                0 => unwrapped_request.name.clone(),
                _ => unwrapped_request.update_name.clone(),
            },
            ..Default::default()
        };

        let db_result = orb_db
            .org_update(&unwrapped_request.name, update_org)
            .expect("There was a problem finding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_remove(
        &self,
        request: Request<OrgRemoveRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        let unwrapped_request = request.into_inner();
        info!("org remove request: {:?}", &unwrapped_request.name);

        let orb_db = postgres::client::OrbitalDBClient::new();

        let db_result = orb_db
            .org_remove(&unwrapped_request.name)
            .expect("There was a problem finding org in database");

        Ok(Response::new(db_result.into()))
    }

    async fn org_list(&self, _request: Request<()>) -> Result<Response<OrgListResponse>, Status> {
        info!("org list request");

        let orb_db = postgres::client::OrbitalDBClient::new();

        // Convert the Vec<Org> response into the proto codegen version.
        let db_result: Vec<OrgEntry> = orb_db
            .org_list()
            .expect("There was a problem finding org in database")
            .into_iter()
            .map(|o| o.into())
            .collect();

        Ok(Response::new(OrgListResponse { orgs: db_result }))
    }
}
