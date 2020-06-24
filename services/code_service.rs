use orbital_headers::code::{
    code_service_server::CodeService, GitRepoAddRequest, GitRepoEntry, GitRepoGetRequest,
    GitRepoListRequest, GitRepoListResponse, GitRepoRemoveRequest, GitRepoUpdateRequest,
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

use log::{debug, info};
use orbital_agent::build_engine;
use orbital_database::postgres;

use serde_json::json;

/// Implementation of protobuf derived `CodeService` trait
#[tonic::async_trait]
impl CodeService for OrbitalApi {
    async fn git_repo_add(
        &self,
        request: Request<GitRepoAddRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        let unwrapped_request = request.into_inner();
        info!("Git repo add: {:?}", &unwrapped_request.name);
        debug!("Git repo add details: {:?}", &unwrapped_request);

        // Declaring this in case we have an ssh key. For test cloning
        let temp_keypath = Temp::new_file().expect("Unable to create temp file");

        // check if repo is public or private. Do a test checkout
        let test_branch = match &unwrapped_request.alt_check_branch.clone().len() {
            0 => "master".to_string(),
            _ => unwrapped_request.alt_check_branch.clone(),
        };

        let (org_db, repo_db, secret_db) = match unwrapped_request.secret_type.clone().into() {
            SecretType::Unspecified => {
                info!("Adding public repo");
                let creds = GitCredentials::Public;

                // Temp dir checkout repo
                match unwrapped_request.skip_check {
                    true => info!("Test git clone check skipped by request"),
                    false => {
                        let _ = match build_engine::clone_repo(
                            &unwrapped_request.uri,
                            &test_branch,
                            creds.clone(),
                        ) {
                            Ok(_) => {
                                info!("Test git clone successful");
                            }
                            Err(_) => {
                                panic!("Test git clone unsuccessful");
                            }
                        };
                    }
                };

                // Write git repo to DB
                let pg_conn = postgres::client::establish_connection();

                // TODO: We should remove usernames from the uri when we add to the database
                // This means we're going to need to add usernames to secret service
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
                info!("Private repo with ssh key");

                let creds = GitCredentials::SshKey {
                    username: unwrapped_request.clone().user,
                    public_key: None,
                    private_key: unwrapped_request.clone().auth_data,
                    passphrase: None,
                };

                // Temp dir checkout repo
                match unwrapped_request.skip_check {
                    true => info!("Test git clone check skipped by request"),
                    false => {
                        let _ = match build_engine::clone_repo(
                            format!(
                                "{}@{}",
                                unwrapped_request.clone().user,
                                &unwrapped_request.uri
                            )
                            .as_str(),
                            &test_branch,
                            creds.clone(),
                        ) {
                            Ok(_) => {
                                info!("Test git clone successful");
                            }
                            Err(_) => {
                                panic!("Test git clone unsuccessful");
                            }
                        };
                    }
                };

                // TODO: We should create a struct for this
                let auth_info = json!({
                    "username": unwrapped_request.clone().user,
                    "private_key": unwrapped_request.clone().auth_data,
                });

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
                    data: auth_info.to_string().into_bytes(),
                });

                debug!("Request for secret add: {:?}", &request);

                let response = secret_client.secret_add(request).await?;
                debug!("RESPONSE = {:?}", response);

                // Convert the response SecretEntry from the secret add into Secret
                let secret: postgres::secret::Secret = response.into_inner().into();

                debug!(
                    "Secret after conversion from proto to DB type: {:?}",
                    &secret
                );

                // Write git repo to DB

                let pg_conn = postgres::client::establish_connection();

                // TODO: Remove username from uri
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
                info!("Private repo with basic auth");
                let creds = GitCredentials::BasicAuth {
                    username: unwrapped_request.clone().user.into(),
                    password: unwrapped_request.clone().auth_data.into(),
                };

                // Temp dir checkout repo
                match unwrapped_request.skip_check {
                    true => info!("Test git clone check skipped by request"),
                    false => {
                        let _ = match build_engine::clone_repo(
                            &unwrapped_request.uri,
                            &test_branch,
                            creds.clone(),
                        ) {
                            Ok(_) => {
                                info!("Test git clone successful");
                            }
                            Err(_) => {
                                panic!("Test git clone unsuccessful");
                            }
                        };
                    }
                };

                // TODO: Create hashmap. Username + password
                let auth_info = json!({
                    "username": unwrapped_request.clone().user,
                    "password": unwrapped_request.clone().auth_data,
                });

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

                debug!("Adding basic auth to secret service");
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
                    data: auth_info.to_string().into_bytes(),
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

                // TODO: Remove username from uri
                postgres::client::repo_add(
                    &pg_conn,
                    &unwrapped_request.org,
                    &unwrapped_request.name,
                    &unwrapped_request.uri,
                    None,
                )
                .expect("There was a problem adding repo in database")
            }
            _ => {
                debug!(
                    "Raw secret type: {:?}",
                    unwrapped_request.secret_type.clone()
                );
                panic!("Only public repo or private repo w/ sshkey/basic auth supported")
            }
        };

        let git_uri_parsed = git_info::git_remote_url_parse(&repo_db.uri.clone()).unwrap();

        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.user.clone().unwrap_or_default(),
            uri: format!("{}", git_uri_parsed.trim_auth()),
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
        let unwrapped_request = request.into_inner();
        info!("Git repo get: {:?}", &unwrapped_request.name);
        debug!("Git repo get details: {:?}", &unwrapped_request);

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

        // Maybe TODO? Call into Secret service to get a username?

        // We use auth_data to hold the name of the secret used by repo
        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            //user: git_uri_parsed.user.unwrap(), // I think we're going to let the build service handle this
            uri: format!("{}", git_uri_parsed.trim_auth()),
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
        let unwrapped_request = request.into_inner();
        info!("Git repo update: {:?}", &unwrapped_request.name);
        debug!("Git repo update details: {:?}", &unwrapped_request);

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
            uri: format!("{}", git_uri_parsed),
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
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.clone().user.unwrap(),
            uri: format!("{}", git_uri_parsed.trim_auth()),
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
        let unwrapped_request = request.into_inner();
        info!("Git repo remove: {:?}", &unwrapped_request.name);
        debug!("Git repo remove details: {:?}", &unwrapped_request);

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
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.clone().user.unwrap_or_default(),
            uri: format!("{}", git_uri_parsed.trim_auth()),
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
        let unwrapped_request = request.into_inner();
        info!("Git repo list request: {:?}", &unwrapped_request);

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
                        git_provider: git_uri_parsed.clone().host.unwrap(),
                        name: repo_db.name,
                        user: git_uri_parsed.clone().user.unwrap_or_default().to_string(),
                        uri: format!("{}", &git_uri_parsed.trim_auth()),
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
