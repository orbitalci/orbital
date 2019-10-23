use tower_grpc::{Request, Response};
use futures::{future, Future, Stream};
use orbital_api::integration::{server, SecretIntegrationEntry, SecretIntegrationListResponse, NotifyIntegrationEntry, NotifyIntegrationListResponse};

#[derive(Clone, Debug)]
struct OrbitalApi;

impl orbital_api::integration::server::IntegrationService for OrbitalApi {
    type CreateSecretIntegrationFuture = future::FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type DeleteSecretIntegrationFuture = future::FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type UpdateSecretIntegrationFuture = future::FutureResult<Response<SecretIntegrationEntry>, tower_grpc::Status>;
    type ListSecretIntegrationsFuture = future::FutureResult<Response<SecretIntegrationListResponse>, tower_grpc::Status>;
    type CreateNotifyIntegrationFuture = future::FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type DeleteNotifyIntegrationFuture = future::FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type UpdateNotifyIntegrationFuture = future::FutureResult<Response<NotifyIntegrationEntry>, tower_grpc::Status>;
    type ListNotifyIntegrationsFuture = future::FutureResult<Response<NotifyIntegrationListResponse>, tower_grpc::Status>;

    fn create_secret_integration(&mut self, request: tower_grpc::Request<orbital_api::integration::SecretIntegrationCreateRequest>) -> Self::CreateSecretIntegrationFuture {
        unimplemented!()
    }

    fn delete_secret_integration(&mut self, request: tower_grpc::Request<orbital_api::integration::SecretIntegrationDeleteRequest>) -> Self::DeleteSecretIntegrationFuture {
        unimplemented!()
    }

    fn update_secret_integration(&mut self, request: tower_grpc::Request<orbital_api::integration::SecretIntegrationUpdateRequest>) -> Self::UpdateSecretIntegrationFuture {
        unimplemented!()
    }

    fn list_secret_integrations(&mut self, request: tower_grpc::Request<orbital_api::integration::SecretIntegrationListRequest>) -> Self::ListSecretIntegrationsFuture {
        unimplemented!()
    }

    fn create_notify_integration(&mut self, request: tower_grpc::Request<orbital_api::integration::NotifyIntegrationCreateRequest>) -> Self::CreateNotifyIntegrationFuture {
        unimplemented!()
    }

    fn delete_notify_integration(&mut self, request: tower_grpc::Request<orbital_api::integration::NotifyIntegrationDeleteRequest>) -> Self::DeleteNotifyIntegrationFuture {
        unimplemented!()
    }

    fn update_notify_integration(&mut self, request: tower_grpc::Request<orbital_api::integration::NotifyIntegrationUpdateRequest>) -> Self::UpdateNotifyIntegrationFuture {
        unimplemented!()
    }

    fn list_notify_integrations(&mut self, request: tower_grpc::Request<orbital_api::integration::NotifyIntegrationListRequest>) -> Self::ListNotifyIntegrationsFuture {
        unimplemented!()
    }
}