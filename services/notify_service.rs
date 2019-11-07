use orbital_headers::notify::{
    notify_service_server::NotifyService, NotifyAddRequest, NotifyEntry, NotifyGetRequest,
    NotifyListRequest, NotifyListResponse, NotifyRemoveRequest, NotifyUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `IntegrationService` trait
#[tonic::async_trait]
impl NotifyService for OrbitalApi {
    async fn notify_add(
        &self,
        _request: Request<NotifyAddRequest>,
    ) -> Result<Response<NotifyEntry>, Status> {
        unimplemented!()
    }

    async fn notify_get(
        &self,
        _request: Request<NotifyGetRequest>,
    ) -> Result<Response<NotifyEntry>, Status> {
        unimplemented!()
    }

    async fn notify_update(
        &self,
        _request: Request<NotifyUpdateRequest>,
    ) -> Result<Response<NotifyEntry>, Status> {
        unimplemented!()
    }

    async fn notify_remove(
        &self,
        _request: Request<NotifyRemoveRequest>,
    ) -> Result<Response<NotifyEntry>, Status> {
        unimplemented!()
    }

    async fn notify_list(
        &self,
        _request: Request<NotifyListRequest>,
    ) -> Result<Response<NotifyListResponse>, Status> {
        unimplemented!()
    }
}
