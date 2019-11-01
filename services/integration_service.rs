use futures::future;
use orbital_headers::integration::{
    NotifyIntegrationEntry, NotifyIntegrationListResponse, SecretIntegrationEntry,
    SecretIntegrationListResponse,
};
use tower_grpc::Response;

#[derive(Clone, Debug)]
struct OrbitalApi;

impl orbital_headers::integration::server::IntegrationService for OrbitalApi {
    type CreateSecretIntegrationFuture =
        future::FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type DeleteSecretIntegrationFuture =
        future::FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type UpdateSecretIntegrationFuture =
        future::FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type ListSecretIntegrationsFuture =
        future::FutureResult<Response<SecretIntegrationListResponse>, tower_grpc::Status>;
    type CreateNotifyIntegrationFuture =
        future::FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type DeleteNotifyIntegrationFuture =
        future::FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type UpdateNotifyIntegrationFuture =
        future::FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type ListNotifyIntegrationsFuture =
        future::FutureResult<Response<NotifyIntegrationListResponse>, tower_grpc::Status>;

    fn create_secret_integration(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::SecretIntegrationCreateRequest>,
    ) -> Self::CreateSecretIntegrationFuture {
        unimplemented!()
    }

    fn delete_secret_integration(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::SecretIntegrationDeleteRequest>,
    ) -> Self::DeleteSecretIntegrationFuture {
        unimplemented!()
    }

    fn update_secret_integration(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::SecretIntegrationUpdateRequest>,
    ) -> Self::UpdateSecretIntegrationFuture {
        unimplemented!()
    }

    fn list_secret_integrations(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::SecretIntegrationListRequest>,
    ) -> Self::ListSecretIntegrationsFuture {
        unimplemented!()
    }

    fn create_notify_integration(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::NotifyIntegrationCreateRequest>,
    ) -> Self::CreateNotifyIntegrationFuture {
        unimplemented!()
    }

    fn delete_notify_integration(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::NotifyIntegrationDeleteRequest>,
    ) -> Self::DeleteNotifyIntegrationFuture {
        unimplemented!()
    }

    fn update_notify_integration(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::NotifyIntegrationUpdateRequest>,
    ) -> Self::UpdateNotifyIntegrationFuture {
        unimplemented!()
    }

    fn list_notify_integrations(
        &mut self,
        _request: tower_grpc::Request<orbital_headers::integration::NotifyIntegrationListRequest>,
    ) -> Self::ListNotifyIntegrationsFuture {
        unimplemented!()
    }
}
