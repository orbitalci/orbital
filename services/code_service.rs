use orbital_headers::code::{
    server::CodeService, GitProviderAddRequest, GitProviderEntry, GitProviderGetRequest,
    GitProviderListRequest, GitProviderListResponse, GitProviderRemoveRequest,
    GitProviderUpdateRequest, GitRepoAddRequest, GitRepoEntry, GitRepoGetRequest,
    GitRepoListRequest, GitRepoListResponse, GitRepoRemoveRequest, GitRepoUpdateRequest,
};

use orbital_headers::secret::{client::SecretServiceClient, SecretAddRequest};

use crate::{OrbitalServiceError, ServiceType};
use orbital_headers::orbital_types::*;

use git_meta::GitCredentials;
use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;

use tonic::{Request, Response, Status};

use super::OrbitalApi;

use agent_runtime::build_engine;
use log::debug;

/// Implementation of protobuf derived `CodeService` trait
#[tonic::async_trait]
impl CodeService for OrbitalApi {
    async fn git_provider_add(
        &self,
        _request: Request<GitProviderAddRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_get(
        &self,
        _request: Request<GitProviderGetRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_update(
        &self,
        _request: Request<GitProviderUpdateRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_remove(
        &self,
        _request: Request<GitProviderRemoveRequest>,
    ) -> Result<Response<GitProviderEntry>, Status> {
        unimplemented!()
    }

    async fn git_provider_list(
        &self,
        _request: Request<GitProviderListRequest>,
    ) -> Result<Response<GitProviderListResponse>, Status> {
        unimplemented!()
    }

    async fn git_repo_add(
        &self,
        request: Request<GitRepoAddRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        debug!("Git repo add request: {:?}", &request);
        let unwrapped_request = request.into_inner();

        // Declaring this in case we have an ssh key. For test cloning
        let temp_keypath = Temp::new_file().expect("Unable to create temp file");

        // check if repo is public or private. Do a test checkout
        let _credentials = match unwrapped_request.secret_type.into() {
            SecretType::Unspecified => {
                debug!("No secret type specified. Public repo");
                let creds = GitCredentials::Public;

                // Temp dir checkout repo
                let _ =
                    match build_engine::clone_repo(&unwrapped_request.uri, "master", creds.clone())
                    {
                        Ok(_) => {
                            debug!("Test git clone successful");
                        }
                        Err(_) => {
                            panic!("Test git clone unsuccessful");
                        }
                    };

                creds
            }
            SecretType::SshKey => {
                debug!("Private repo with ssh key");

                // Write private key into a temp file
                debug!("Writing incoming ssh key to temp file");
                let mut file = File::create(temp_keypath.as_path())?;
                let mut _contents = String::new();
                let _ = file.write_all(unwrapped_request.auth_data.as_bytes());

                let creds = GitCredentials::SshKey {
                    username: "git".into(),
                    public_key: None,
                    private_key: temp_keypath.as_path(),
                    passphrase: None,
                };

                // Temp dir checkout repo
                debug!("Test cloning the repo");
                let _ =
                    match build_engine::clone_repo(&unwrapped_request.uri, "master", creds.clone())
                    {
                        Ok(_) => {
                            debug!("Test git clone successful");
                        }
                        Err(_) => {
                            panic!("Test git clone unsuccessful");
                        }
                    };

                // Commit secret into secret service
                debug!("Connecting to the Secret service");
                let secret_client_conn_req = SecretServiceClient::connect(format!(
                    "http://{}",
                    super::get_service_uri(ServiceType::Secret)
                ));
                let mut secret_client = match secret_client_conn_req.await {
                    Ok(connection_handler) => connection_handler,
                    Err(_e) => {
                        return Err(
                            OrbitalServiceError::new("Unable to connect to Secret service").into(),
                        )
                    }
                };

                debug!("Adding private key to secret service");
                let org_name = &unwrapped_request.org;
                let secret_name = format!("{}", &unwrapped_request.name);
                let vault_path = format!(
                    "orbital/{}/{}/{}",
                    org_name,
                    SecretType::from(unwrapped_request.secret_type),
                    secret_name,
                )
                .to_lowercase();

                let request = Request::new(SecretAddRequest {
                    org: org_name.into(),
                    name: vault_path.into(),
                    secret_type: SecretType::from(unwrapped_request.secret_type).into(),
                    data: unwrapped_request.auth_data.into(),
                });

                debug!("Request for secret add: {:?}", &request);

                let response = secret_client.secret_add(request).await?;
                println!("RESPONSE = {:?}", response);

                creds
            }
            SecretType::BasicAuth => {
                debug!("Private repo with basic auth");
                let creds = GitCredentials::UserPassPlaintext {
                    username: "git".to_string(),
                    password: "fakepassword".to_string(),
                };

                // Temp dir checkout repo
                debug!("Test cloning the repo");
                let _ =
                    match build_engine::clone_repo(&unwrapped_request.uri, "master", creds.clone())
                    {
                        Ok(_) => {
                            debug!("Test git clone successful");
                        }
                        Err(_) => {
                            panic!("Test git clone unsuccessful");
                        }
                    };

                creds
            }
            _ => panic!("Only public repo or private repo w/ sshkey/basic auth supported"),
        };

        // Test checking code out before committing repo into database and vault

        let response = Response::new(GitRepoEntry {
            ..Default::default()
        });

        Ok(response)
    }

    async fn git_repo_get(
        &self,
        request: Request<GitRepoGetRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        debug!("Git repo get request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // Connect to database. Query for the repo
        // let git_repo = db_get_repo(org, git_provider, git_repo_name)-> GitRepoEntry

        let mut git_repo = GitRepoEntry::default();
        git_repo.org = unwrapped_request.org;
        git_repo.user = "git".into();
        git_repo.git_provider = unwrapped_request.git_provider;
        git_repo.name = unwrapped_request.name;
        git_repo.secret_type = SecretType::SshKey.into();
        git_repo.uri = unwrapped_request.uri;
        git_repo.auth_data =
            "secret/orbital/default_org/sshkey/github.com/level11consulting/orbitalci".into();

        debug!("Response: {:?}", &git_repo);
        Ok(Response::new(git_repo))
    }

    async fn git_repo_update(
        &self,
        _request: Request<GitRepoUpdateRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        unimplemented!()
    }

    async fn git_repo_remove(
        &self,
        _request: Request<GitRepoRemoveRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        unimplemented!()
    }

    async fn git_repo_list(
        &self,
        _request: Request<GitRepoListRequest>,
    ) -> Result<Response<GitRepoListResponse>, Status> {
        unimplemented!()
    }
}
