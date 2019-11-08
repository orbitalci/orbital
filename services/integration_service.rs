use orbital_headers::integration::{
    server::IntegrationService, NotifyIntegrationCreateRequest, NotifyIntegrationDeleteRequest,
    NotifyIntegrationEntry, NotifyIntegrationListRequest, NotifyIntegrationListResponse,
    NotifyIntegrationUpdateRequest, SecretIntegrationCreateRequest, SecretIntegrationDeleteRequest,
    SecretIntegrationEntry, SecretIntegrationListRequest, SecretIntegrationListResponse,
    SecretIntegrationUpdateRequest,
};
use tonic::{Request, Response, Status};

use super::OrbitalApi;

/// Implementation of protobuf derived `IntegrationService` trait
#[tonic::async_trait]
impl IntegrationService for OrbitalApi {
    async fn create_secret_integration(
        &self,
        _request: Request<SecretIntegrationCreateRequest>,
    ) -> Result<Response<SecretIntegrationEntry>, Status> {
        unimplemented!()
    }

    async fn delete_secret_integration(
        &self,
        _request: Request<SecretIntegrationDeleteRequest>,
    ) -> Result<Response<SecretIntegrationEntry>, Status> {
        unimplemented!()
    }

    async fn update_secret_integration(
        &self,
        _request: Request<SecretIntegrationUpdateRequest>,
    ) -> Result<Response<SecretIntegrationEntry>, Status> {
        unimplemented!()
    }

    async fn list_secret_integrations(
        &self,
        _request: Request<SecretIntegrationListRequest>,
    ) -> Result<Response<SecretIntegrationListResponse>, Status> {
        unimplemented!()
    }

    async fn create_notify_integration(
        &self,
        _request: Request<NotifyIntegrationCreateRequest>,
    ) -> Result<Response<NotifyIntegrationEntry>, Status> {
        unimplemented!()
    }

    async fn delete_notify_integration(
        &self,
        _request: Request<NotifyIntegrationDeleteRequest>,
    ) -> Result<Response<NotifyIntegrationEntry>, Status> {
        unimplemented!()
    }

    async fn update_notify_integration(
        &self,
        _request: Request<NotifyIntegrationUpdateRequest>,
    ) -> Result<Response<NotifyIntegrationEntry>, Status> {
        unimplemented!()
    }

    async fn list_notify_integrations(
        &self,
        _request: Request<NotifyIntegrationListRequest>,
    ) -> Result<Response<NotifyIntegrationListResponse>, Status> {
        unimplemented!()
    }
}
