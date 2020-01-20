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

                // Write git repo to DB

                let pg_conn = postgres::client::establish_connection();

                let _db_result = postgres::client::repo_add(
                    &pg_conn,
                    &unwrapped_request.org,
                    &unwrapped_request.name,
                    &unwrapped_request.uri,
                    Some(secret),
                )
                .expect("There was a problem adding repo in database");

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
        let pg_conn = postgres::client::establish_connection();

        let db_result =
            //postgres::client::repo_get(&pg_conn, &unwrapped_request.org, &unwrapped_request.name)
            postgres::client::repo_get(&pg_conn, &unwrapped_request.org, &format!("{}/{}", &unwrapped_request.org, &unwrapped_request.name))
                .expect("There was a problem getting repo in database");

        let git_uri_parsed = git_info::git_remote_url_parse(&db_result.uri.clone());

        let vault_path = match db_result.secret_id {
            Some(id) => {
                postgres::client::secret_from_id(&pg_conn, id)
                    .expect("Couldn't resolve secret id to a secret")
                    .vault_path
            }
            None => "".to_string(),
        };

        let mut git_repo = GitRepoEntry::default();
        git_repo.org = postgres::client::org_from_id(&pg_conn, db_result.org_id)
            .expect("Couldn't resolve org id to a name")
            .name;
        git_repo.user = git_uri_parsed.user;
        git_repo.git_provider = git_uri_parsed.provider;
        git_repo.name = git_uri_parsed.repo;
        git_repo.secret_type =
            postgres::client::secret_from_id(&pg_conn, db_result.secret_id.unwrap_or(0))
                .unwrap_or(postgres::secret::Secret::default())
                .secret_type
                .into();
        git_repo.uri = db_result.uri;
        git_repo.auth_data = vault_path;

        debug!("Response: {:?}", &git_repo);
        Ok(Response::new(git_repo))
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

        // FIXME: We're not really doing a whole lot to support updating secrets at the moment
        // Build the NewRepo struct
        let update_repo = postgres::repo::NewRepo {
            org_id: current_repo.org_id,
            name: current_repo.name,
            uri: unwrapped_request.uri,
            git_host_type: current_repo.git_host_type,
            secret_id: current_repo.secret_id,
            build_active_state: current_repo.build_active_state,
            notify_active_state: current_repo.notify_active_state,
            next_build_index: current_repo.next_build_index,
        };

        let db_result = postgres::client::repo_update(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.name,
            update_repo,
        )
        .expect("Could not update repo");

        let git_uri_parsed = git_info::git_remote_url_parse(&db_result.uri.clone());

        let vault_path = match db_result.secret_id {
            Some(id) => {
                postgres::client::secret_from_id(&pg_conn, id)
                    .expect("Couldn't resolve secret id to a secret")
                    .vault_path
            }
            None => "".to_string(),
        };

        let mut git_repo = GitRepoEntry::default();
        git_repo.org = unwrapped_request.org;
        git_repo.user = git_uri_parsed.user;
        git_repo.git_provider = git_uri_parsed.provider;
        git_repo.name = git_uri_parsed.repo;
        git_repo.secret_type =
            postgres::client::secret_from_id(&pg_conn, db_result.secret_id.unwrap_or(0))
                .unwrap_or(postgres::secret::Secret::default())
                .secret_type
                .into();
        git_repo.uri = db_result.uri;
        git_repo.auth_data = vault_path;

        debug!("Response: {:?}", &git_repo);
        Ok(Response::new(git_repo))
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

        let git_uri_parsed = git_info::git_remote_url_parse(&db_result.uri.clone());

        let vault_path = match db_result.secret_id {
            Some(id) => {
                postgres::client::secret_from_id(&pg_conn, id)
                    .expect("Couldn't resolve secret id to a secret")
                    .vault_path
            }
            None => "".to_string(),
        };

        let mut git_repo = GitRepoEntry::default();
        git_repo.org = unwrapped_request.org;
        git_repo.user = git_uri_parsed.user;
        git_repo.git_provider = git_uri_parsed.provider;
        git_repo.name = git_uri_parsed.repo;
        git_repo.secret_type =
            postgres::client::secret_from_id(&pg_conn, db_result.secret_id.unwrap_or(0))
                .unwrap_or(postgres::secret::Secret::default())
                .secret_type
                .into();
        git_repo.uri = db_result.uri;
        git_repo.auth_data = vault_path;

        debug!("Response: {:?}", &git_repo);
        Ok(Response::new(git_repo))
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
                .map(|o| o.into())
                .collect();

        let mut git_repos = GitRepoListResponse::default();
        git_repos.git_repos = db_result;
        debug!("Response: {:?}", &git_repos);
        Ok(Response::new(git_repos))
    }
}
