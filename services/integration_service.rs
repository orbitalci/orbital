use orbital_headers::integration::{
    server::IntegrationService, NotifyCreateRequest, NotifyDeleteRequest, NotifyEntry,
    NotifyListRequest, NotifyListResponse, NotifyUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `IntegrationService` trait
#[tonic::async_trait]
impl IntegrationService for OrbitalApi {
    async fn create_notify(
        &self,
        _request: Request<NotifyCreateRequest>,
    ) -> Result<Response<NotifyEntry>, Status> {
        unimplemented!()
    }

    async fn delete_notify(
        &self,
        _request: Request<NotifyDeleteRequest>,
    ) -> Result<Response<NotifyEntry>, Status> {
        unimplemented!()
    }

    async fn update_notify(
        &self,
        _request: Request<NotifyUpdateRequest>,
    ) -> Result<Response<NotifyEntry>, Status> {
        unimplemented!()
    }

    async fn list_notify(
        &self,
        _request: Request<NotifyListRequest>,
    ) -> Result<Response<NotifyListResponse>, Status> {
        unimplemented!()
    }
}
