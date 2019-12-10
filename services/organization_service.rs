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

        let db_record = postgres::org::NewOrg {
            name: unwrapped_request.name,
        };

        let db_result = postgres::client::new_org(&pg_conn, db_record);

        Ok(Response::new(OrgEntry {
            name: db_result.name.into(),
            active_state: true.into(),
        }))
    }

    async fn org_get(
        &self,
        _request: Request<OrgGetRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        unimplemented!()
    }

    async fn org_update(
        &self,
        _request: Request<OrgUpdateRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        unimplemented!()
    }

    async fn org_remove(
        &self,
        _request: Request<OrgRemoveRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        unimplemented!()
    }

    async fn org_list(&self, _request: Request<()>) -> Result<Response<OrgListResponse>, Status> {
        unimplemented!()
    }
}
