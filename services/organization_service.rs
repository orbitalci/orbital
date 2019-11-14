use orbital_headers::organization::{
    server::OrganizationService, OrgAddRequest, OrgEntry, OrgGetRequest, OrgListResponse,
    OrgRemoveRequest, OrgUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `OrganizationService` trait
#[tonic::async_trait]
impl OrganizationService for OrbitalApi {
    async fn org_add(
        &self,
        _request: Request<OrgAddRequest>,
    ) -> Result<Response<OrgEntry>, Status> {
        unimplemented!()
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
