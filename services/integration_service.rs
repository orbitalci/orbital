use futures::future::FutureResult;
use orbital_headers::integration::{
    server::IntegrationService, NotifyIntegrationCreateRequest, NotifyIntegrationDeleteRequest,
    NotifyIntegrationEntry, NotifyIntegrationListRequest, NotifyIntegrationListResponse,
    NotifyIntegrationUpdateRequest, SecretIntegrationCreateRequest, SecretIntegrationDeleteRequest,
    SecretIntegrationEntry, SecretIntegrationListRequest, SecretIntegrationListResponse,
    SecretIntegrationUpdateRequest,
};
use tower_grpc::Response;

use super::OrbitalApi;

/// Implementation of protobuf derived `IntegrationService` trait
impl IntegrationService for OrbitalApi {
    type CreateSecretIntegrationFuture =
        FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type DeleteSecretIntegrationFuture =
        FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type UpdateSecretIntegrationFuture =
        FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type ListSecretIntegrationsFuture =
        FutureResult<Response<SecretIntegrationListResponse>, tower_grpc::Status>;
    type CreateNotifyIntegrationFuture =
        FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type DeleteNotifyIntegrationFuture =
        FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type UpdateNotifyIntegrationFuture =
        FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type ListNotifyIntegrationsFuture =
        FutureResult<Response<NotifyIntegrationListResponse>, tower_grpc::Status>;

    fn create_secret_integration(
        &mut self,
        _request: tower_grpc::Request<SecretIntegrationCreateRequest>,
    ) -> Self::CreateSecretIntegrationFuture {
        unimplemented!()
    }

    fn delete_secret_integration(
        &mut self,
        _request: tower_grpc::Request<SecretIntegrationDeleteRequest>,
    ) -> Self::DeleteSecretIntegrationFuture {
        unimplemented!()
    }

    fn update_secret_integration(
        &mut self,
        _request: tower_grpc::Request<SecretIntegrationUpdateRequest>,
    ) -> Self::UpdateSecretIntegrationFuture {
        unimplemented!()
    }

    fn list_secret_integrations(
        &mut self,
        _request: tower_grpc::Request<SecretIntegrationListRequest>,
    ) -> Self::ListSecretIntegrationsFuture {
        unimplemented!()
    }

    fn create_notify_integration(
        &mut self,
        _request: tower_grpc::Request<NotifyIntegrationCreateRequest>,
    ) -> Self::CreateNotifyIntegrationFuture {
        unimplemented!()
    }

    fn delete_notify_integration(
        &mut self,
        _request: tower_grpc::Request<NotifyIntegrationDeleteRequest>,
    ) -> Self::DeleteNotifyIntegrationFuture {
        unimplemented!()
    }

    fn update_notify_integration(
        &mut self,
        _request: tower_grpc::Request<NotifyIntegrationUpdateRequest>,
    ) -> Self::UpdateNotifyIntegrationFuture {
        unimplemented!()
    }

    fn list_notify_integrations(
        &mut self,
        _request: tower_grpc::Request<NotifyIntegrationListRequest>,
    ) -> Self::ListNotifyIntegrationsFuture {
        unimplemented!()
    }
}
