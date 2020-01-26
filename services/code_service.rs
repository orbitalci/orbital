use orbital_headers::code::{
    code_service_server::CodeService, GitProviderAddRequest, GitProviderEntry,
    GitProviderGetRequest, GitProviderListRequest, GitProviderListResponse,
    GitProviderRemoveRequest, GitProviderUpdateRequest, GitRepoAddRequest, GitRepoEntry,
    GitRepoGetRequest, GitRepoListRequest, GitRepoListResponse, GitRepoRemoveRequest,
    GitRepoUpdateRequest,
};

use orbital_headers::secret::{secret_service_client::SecretServiceClient, SecretAddRequest};

use crate::{OrbitalServiceError, ServiceType};
use orbital_headers::orbital_types::*;

use git_meta::git_info;
use git_meta::GitCredentials;
use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;

use tonic::{Request, Response, Status};

use super::OrbitalApi;

use agent_runtime::build_engine;
use log::debug;
use orbital_database::postgres;

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

        let (org_db, repo_db, secret_db) = match unwrapped_request.secret_type.into() {
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

                // Write git repo to DB
                let pg_conn = postgres::client::establish_connection();

                postgres::client::repo_add(
                    &pg_conn,
                    &unwrapped_request.org,
                    &unwrapped_request.name,
                    &unwrapped_request.uri,
                    None,
                )
                .expect("There was a problem adding repo in database")
            }
            SecretType::SshKey => {
                debug!("Private repo with ssh key");

                // Write private key into a temp file
                debug!("Writing incoming ssh key to temp file");
                let mut file = File::create(temp_keypath.as_path())?;
                let mut _contents = String::new();
                let _ = file.write_all(unwrapped_request.auth_data.as_bytes());

                let creds = GitCredentials::SshKey {
                    username: unwrapped_request.user,
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
                let secret_name = format!(
                    "{}/{}",
                    &unwrapped_request.git_provider, &unwrapped_request.name
                );

                let request = Request::new(SecretAddRequest {
                    org: org_name.into(),
                    name: secret_name.into(),
                    secret_type: SecretType::from(unwrapped_request.secret_type).into(),
                    data: unwrapped_request.auth_data.into(),
                });

                debug!("Request for secret add: {:?}", &request);

                let response = secret_client.secret_add(request).await?;
                println!("RESPONSE = {:?}", response);

                // Convert the response SecretEntry from the secret add into Secret
                let secret: postgres::secret::Secret = response.into_inner().into();

                debug!(
                    "Secret after conversion from proto to DB type: {:?}",
                    &secret
                );

                // Write git repo to DB

                let pg_conn = postgres::client::establish_connection();

                postgres::client::repo_add(
                    &pg_conn,
                    &unwrapped_request.org,
                    &unwrapped_request.name,
                    &unwrapped_request.uri,
                    Some(secret.clone()),
                )
                .expect("There was a problem adding repo in database")
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

                debug!("Adding userpass to secret service");
                let org_name = &unwrapped_request.org;
                let secret_name = format!(
                    "{}/{}",
                    &unwrapped_request.git_provider, &unwrapped_request.name
                );

                // TODO: auth_data does not have a representation for userpass
                let request = Request::new(SecretAddRequest {
                    org: org_name.into(),
                    name: secret_name.into(),
                    secret_type: SecretType::from(unwrapped_request.secret_type).into(),
                    data: unwrapped_request.auth_data.into(),
                });

                debug!("Request for secret add: {:?}", &request);

                let response = secret_client.secret_add(request).await?;
                println!("RESPONSE = {:?}", response);

                // Convert the response SecretEntry from the secret add into Secret
                let secret: postgres::secret::Secret = response.into_inner().into();

                debug!(
                    "Secret after conversion from proto to DB type: {:?}",
                    &secret
                );

                // Write git repo to DB

                let pg_conn = postgres::client::establish_connection();

                postgres::client::repo_add(
                    &pg_conn,
                    &unwrapped_request.org,
                    &unwrapped_request.name,
                    &unwrapped_request.uri,
                    None,
                )
                .expect("There was a problem adding repo in database")
            }
            _ => panic!("Only public repo or private repo w/ sshkey/basic auth supported"),
        };

        let git_uri_parsed = git_info::git_remote_url_parse(&repo_db.uri.clone()).unwrap();

        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.user.unwrap(),
            uri: git_uri_parsed.href,
            secret_type: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .secret_type
                .into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index.into(),
            auth_data: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .name
                .into(),
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
        let pg_conn = postgres::client::establish_connection();

        let db_result =
            postgres::client::repo_get(&pg_conn, &unwrapped_request.org, &unwrapped_request.name)
                .expect("There was a problem getting repo in database");

        debug!("repo get db result: {:?}", &db_result);

        let org_db = db_result.0;
        let repo_db = db_result.1;
        let secret_db = db_result.2;

        let git_uri_parsed = git_info::git_remote_url_parse(&repo_db.uri.clone()).unwrap();

        // We use auth_data to hold the name of the secret used by repo
        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.user.unwrap(),
            uri: git_uri_parsed.href,
            secret_type: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .secret_type
                .into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index.into(),
            auth_data: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .name
                .into(),
            ..Default::default()
        });

        debug!("Response: {:?}", &response);
        Ok(response)
    }

    async fn git_repo_update(
        &self,
        request: Request<GitRepoUpdateRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        debug!("Git repo update request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        // Get the current repo
        let current_repo =
            postgres::client::repo_get(&pg_conn, &unwrapped_request.org, &unwrapped_request.name)
                .expect("Could not find repo to update");

        let org_db = current_repo.0;
        let repo_db = current_repo.1;
        let secret_db = current_repo.2;

        let secret_id = match &secret_db {
            Some(s) => Some(s.id),
            None => None,
        };

        let git_uri_parsed = git_info::git_remote_url_parse(&repo_db.uri.clone()).unwrap();

        // FIXME: We're not really doing a whole lot to support updating secrets at the moment
        // Build the NewRepo struct
        let update_repo = postgres::repo::NewRepo {
            org_id: org_db.id,
            name: repo_db.name,
            uri: git_uri_parsed.href,
            git_host_type: repo_db.git_host_type,
            secret_id: secret_id,
            build_active_state: repo_db.build_active_state,
            notify_active_state: repo_db.notify_active_state,
            next_build_index: repo_db.next_build_index,
        };

        let db_result = postgres::client::repo_update(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
            update_repo,
        )
        .expect("Could not update repo");

        // Repeating ourselves to write response with update data
        let org_db = db_result.0;
        let repo_db = db_result.1;
        let secret_db = db_result.2;

        let git_uri_parsed = git_info::git_remote_url_parse(&repo_db.uri.clone()).unwrap();

        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.user.unwrap(),
            uri: git_uri_parsed.href,
            secret_type: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .secret_type
                .into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index.into(),
            auth_data: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .name
                .into(),
            ..Default::default()
        });

        debug!("Response: {:?}", &response);
        Ok(response)
    }

    async fn git_repo_remove(
        &self,
        request: Request<GitRepoRemoveRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        debug!("Git repo remove request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        let db_result = postgres::client::repo_remove(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
        )
        .expect("There was a problem removing repo in database");

        debug!("repo remove db result: {:?}", &db_result);

        let org_db = db_result.0;
        let repo_db = db_result.1;
        let secret_db = db_result.2;

        let git_uri_parsed = git_info::git_remote_url_parse(&repo_db.uri.clone()).unwrap();

        // We use auth_data to hold the name of the secret used by repo
        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.user.unwrap(),
            uri: git_uri_parsed.href,
            secret_type: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .secret_type
                .into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index.into(),
            auth_data: secret_db
                .clone()
                .unwrap_or(postgres::secret::Secret::default())
                .name
                .into(),
            ..Default::default()
        });

        debug!("Response: {:?}", &response);
        Ok(response)
    }

    async fn git_repo_list(
        &self,
        request: Request<GitRepoListRequest>,
    ) -> Result<Response<GitRepoListResponse>, Status> {
        debug!("Git repo list request: {:?}", &request);

        let unwrapped_request = request.into_inner();

        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        let db_result: Vec<GitRepoEntry> =
            postgres::client::repo_list(&pg_conn, &unwrapped_request.org)
                .expect("There was a problem listing repo from database")
                .into_iter()
                .map(|(org_db, repo_db, secret_db)| {
                    let git_uri_parsed =
                        git_info::git_remote_url_parse(&repo_db.uri.clone()).unwrap();
                    GitRepoEntry {
                        org: org_db.name,
                        git_provider: git_uri_parsed.host.unwrap(),
                        name: repo_db.name,
                        user: git_uri_parsed.user.unwrap(),
                        uri: git_uri_parsed.href,
                        secret_type: secret_db
                            .clone()
                            .unwrap_or(postgres::secret::Secret::default())
                            .secret_type
                            .into(),
                        build: repo_db.build_active_state.into(),
                        notify: repo_db.notify_active_state.into(),
                        next_build_index: repo_db.next_build_index.into(),
                        auth_data: secret_db
                            .clone()
                            .unwrap_or(postgres::secret::Secret::default())
                            .name
                            .into(),
                        ..Default::default()
                    }
                })
                .collect();

        let mut git_repos = GitRepoListResponse::default();
        git_repos.git_repos = db_result;
        debug!("Response: {:?}", &git_repos);
        Ok(Response::new(git_repos))
    }
}
