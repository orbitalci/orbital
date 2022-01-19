use crate::orbital_headers::code::{
    code_service_server::CodeService, GitRepoAddRequest, GitRepoEntry, GitRepoGetRequest,
    GitRepoListRequest, GitRepoListResponse, GitRepoRemoteBranchHead, GitRepoRemoteBranchHeadList,
    GitRepoRemoveRequest, GitRepoUpdateRequest,
};

use crate::orbital_headers::secret::{
    secret_service_client::SecretServiceClient, SecretAddRequest,
};

use crate::orbital_headers::orbital_types::*;
use crate::orbital_services::{OrbitalServiceError, ServiceType};

use git_meta::GitCredentials;
use git_url_parse::GitUrl;
use mktemp::Temp;

use tonic::{Request, Response, Status};

use std::fs::File;
use std::io::prelude::*;

use super::OrbitalApi;

use crate::orbital_database::postgres;
use crate::orbital_utils::orbital_agent::build_engine;
use log::{debug, info};

use log::warn;
use serde_json::json;

/// Implementation of protobuf derived `CodeService` trait
#[tonic::async_trait]
impl CodeService for OrbitalApi {
    async fn git_repo_add(
        &self,
        request: Request<GitRepoAddRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        let req = request.into_inner();
        info!("Git repo add: {:?}", &req.name);
        debug!("Git repo add details: {:?}", &req);

        // We want to check out the default branch if not specified.
        // check if repo is public or private. Do a test checkout
        let branch = match &req.alt_check_branch.clone().len() {
            0 => None,
            _ => Some(req.alt_check_branch.clone()),
        };

        // We want to convert the list of branch refs into serde_json for postgres
        let mut branch_heads = serde_json::Map::new();
        if let Some(branch_ref_proto) = &req.remote_branch_heads {
            for proto in branch_ref_proto.remote_branch_heads.clone() {
                branch_heads.insert(proto.branch, serde_json::Value::String(proto.commit));
            }
        }

        // Create a temp file for writing private key, just in case
        let private_key = Temp::new_file().expect("Unable to create temp file for private key");

        let (org_db, repo_db, secret_db) = match req.secret_type.into() {
            SecretType::Unspecified => {
                info!("Adding public repo");
                let creds = None;

                // Temp dir checkout repo
                match req.skip_check {
                    true => info!("Test git clone check skipped by request"),
                    false => {
                        let temp_dir = Temp::new_dir().expect("Creating test clone dir failed");

                        let _ = match build_engine::clone_repo(
                            &req.uri,
                            branch.as_ref(),
                            creds,
                            temp_dir.as_path(),
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
                let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(req.org));

                // TODO: We should remove usernames from the uri when we add to the database
                // This means we're going to need to add usernames to secret service
                orb_db
                    .repo_add(
                        &req.name,
                        &req.uri,
                        &req.canonical_branch,
                        None,
                        json!(branch_heads),
                    )
                    .expect("There was a problem adding repo in database")
            }
            SecretType::SshKey => {
                info!("Private repo with ssh key");

                let mut file = match File::create(&private_key) {
                    Err(why) => panic!("couldn't create {}: {}", &private_key.display(), why),
                    Ok(file) => file,
                };

                match file.write_all(req.clone().auth_data.as_bytes()) {
                    Err(why) => {
                        panic!("couldn't write to {}: {}", &private_key.display(), why)
                    }
                    Ok(_) => println!("successfully wrote to {}", &private_key.display()),
                }

                let creds = Some(GitCredentials::SshKey {
                    username: req.clone().user,
                    public_key: None,
                    private_key: private_key.to_path_buf(),
                    passphrase: None,
                });

                // Temp dir checkout repo
                match req.skip_check {
                    true => info!("Test git clone check skipped by request"),
                    false => {
                        let temp_dir = Temp::new_dir().expect("Creating test clone dir failed");

                        let _ = match build_engine::clone_repo(
                            format!("{}@{}", req.clone().user, &req.uri).as_str(),
                            branch.as_deref(),
                            creds.clone(),
                            temp_dir.as_path(),
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
                    "username": req.clone().user,
                    "private_key": req.clone().auth_data,
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
                let org_name = &req.org;
                let secret_name = format!("{}/{}", &req.git_provider, &req.name);

                let request = Request::new(SecretAddRequest {
                    org: org_name.into(),
                    name: secret_name,
                    secret_type: SecretType::from(req.secret_type).into(),
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

                let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(req.org));

                // TODO: Remove username from uri
                orb_db
                    .repo_add(
                        &req.name,
                        &req.uri,
                        &req.canonical_branch,
                        Some(secret),
                        json!(branch_heads),
                    )
                    .expect("There was a problem adding repo in database")
            }
            SecretType::BasicAuth => {
                info!("Private repo with basic auth");
                let creds = Some(GitCredentials::UserPassPlaintext {
                    username: req.clone().user,
                    password: req.clone().auth_data,
                });

                // Temp dir checkout repo
                match req.skip_check {
                    true => info!("Test git clone check skipped by request"),
                    false => {
                        let temp_dir = Temp::new_dir()?;

                        let _ = match build_engine::clone_repo(
                            &req.uri,
                            branch.as_ref(),
                            creds.clone(),
                            temp_dir.as_ref(),
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
                    "username": req.clone().user,
                    "password": req.clone().auth_data,
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
                let org_name = &req.org;
                let secret_name = format!("{}/{}", &req.git_provider, &req.name);

                // TODO: auth_data does not have a representation for userpass
                let request = Request::new(SecretAddRequest {
                    org: org_name.into(),
                    name: secret_name,
                    secret_type: SecretType::from(req.secret_type).into(),
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

                let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(req.org));

                // TODO: Remove username from uri
                orb_db
                    .repo_add(
                        &req.name,
                        &req.uri,
                        &req.canonical_branch,
                        None,
                        json!(branch_heads),
                    )
                    .expect("There was a problem adding repo in database")
            }
            _ => {
                debug!("Raw secret type: {:?}", req.secret_type.clone());
                panic!("Only public repo or private repo w/ sshkey/basic auth supported")
            }
        };

        let git_uri_parsed = GitUrl::parse(&repo_db.uri).unwrap();

        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.user.clone().unwrap_or_default(),
            uri: format!("{}", git_uri_parsed.trim_auth()),
            canonical_branch: repo_db.canonical_branch,
            secret_type: secret_db.clone().unwrap_or_default().secret_type.into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index,
            auth_data: secret_db.unwrap_or_default().name,
            remote_branch_heads: None,
        });

        Ok(response)
    }

    async fn git_repo_get(
        &self,
        request: Request<GitRepoGetRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        let req = request.into_inner();
        info!("Git repo get: {:?}", &req.name);
        debug!("Git repo get details: {:?}", &req);

        // Connect to database. Query for the repo
        let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(req.org));

        let db_result = orb_db
            .repo_get(&req.name)
            .expect("There was a problem getting repo in database");

        debug!("repo get db result: {:?}", &db_result);

        let org_db = db_result.0;
        let repo_db = db_result.1;
        let secret_db = db_result.2;

        let git_uri_parsed = GitUrl::parse(&repo_db.uri).unwrap();

        // Maybe TODO? Call into Secret service to get a username?

        // Convert branch_heads to proto type GitRepoRemoteBranchHeadList
        let mut branch_head_proto_list = GitRepoRemoteBranchHeadList {
            remote_branch_heads: Vec::new(),
        };
        for (k, v) in repo_db.remote_branch_heads.as_object().unwrap().iter() {
            let proto_branch = GitRepoRemoteBranchHead {
                branch: k.to_string(),
                commit: v.as_str().unwrap().to_string(),
            };

            branch_head_proto_list
                .remote_branch_heads
                .push(proto_branch);
        }

        // We use auth_data to hold the name of the secret used by repo
        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            //user: git_uri_parsed.user.unwrap(), // I think we're going to let the build service handle this
            uri: format!("{}", git_uri_parsed.trim_auth()),
            secret_type: secret_db.clone().unwrap_or_default().secret_type.into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index,
            auth_data: secret_db.unwrap_or_default().name,
            remote_branch_heads: Some(branch_head_proto_list),
            ..Default::default()
        });

        debug!("Response: {:?}", &response);
        Ok(response)
    }

    async fn git_repo_update(
        &self,
        request: Request<GitRepoUpdateRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        let req = request.into_inner();
        info!("Git repo update: {:?}", &req.name);
        debug!("Git repo update details: {:?}", &req);

        // Connect to database. Query for the repo
        let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(req.org));

        // Get the current repo
        let current_repo = orb_db
            .repo_get(&req.name)
            .expect("Could not find repo to update");

        let org_db = current_repo.0;
        let repo_db = current_repo.1;
        let secret_db = current_repo.2;

        let secret_id = secret_db.as_ref().map(|s| s.id);

        let git_uri_parsed = GitUrl::parse(&repo_db.uri).unwrap();

        // We want to convert the list of branch refs into serde_json for postgres
        let mut branch_heads = serde_json::Map::new();
        if let Some(branch_ref_proto) = &req.remote_branch_heads {
            for proto in branch_ref_proto.remote_branch_heads.clone() {
                branch_heads.insert(proto.branch, serde_json::Value::String(proto.commit));
            }
        }

        // FIXME: We're not really doing a whole lot to support updating secrets at the moment
        // Build the NewRepo struct
        let update_repo = postgres::repo::NewRepo {
            org_id: org_db.id,
            name: repo_db.name,
            uri: format!("{}", git_uri_parsed),
            canonical_branch: repo_db.canonical_branch,
            git_host_type: repo_db.git_host_type,
            secret_id,
            build_active_state: repo_db.build_active_state,
            notify_active_state: repo_db.notify_active_state,
            next_build_index: repo_db.next_build_index,
            // TODO
            remote_branch_heads: json!(branch_heads),
        };

        let db_result = orb_db
            .repo_update(&req.name, update_repo)
            .expect("Could not update repo");

        // Repeating ourselves to write response with update data
        let org_db = db_result.0;
        let repo_db = db_result.1;
        let secret_db = db_result.2;

        let git_uri_parsed = GitUrl::parse(&repo_db.uri).unwrap();

        // TODO: Build a return value from db_result for branch_heads
        //let branch_heads_db = None;

        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.clone().user.unwrap_or_default(),
            uri: format!("{}", git_uri_parsed.trim_auth()),
            secret_type: secret_db.clone().unwrap_or_default().secret_type.into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index,
            auth_data: secret_db.unwrap_or_default().name,
            remote_branch_heads: {
                match repo_db.remote_branch_heads {
                    serde_json::Value::Null => None,
                    serde_json::Value::Object(map_value) => {
                        let mut git_branches: Vec<GitRepoRemoteBranchHead> = Vec::new();

                        for (k, v) in map_value {
                            let branch = GitRepoRemoteBranchHead {
                                branch: k,
                                commit: v.to_string(),
                            };

                            git_branches.push(branch);
                        }
                        Some(GitRepoRemoteBranchHeadList {
                            remote_branch_heads: git_branches,
                        })
                    }
                    _ => {
                        warn!("There was a serde Value other than an Object. Dropping value. ");
                        None
                    }
                }
            },
            ..Default::default()
        });

        debug!("Response: {:?}", &response);
        Ok(response)
    }

    async fn git_repo_remove(
        &self,
        request: Request<GitRepoRemoveRequest>,
    ) -> Result<Response<GitRepoEntry>, Status> {
        let req = request.into_inner();
        info!("Git repo remove: {:?}", &req.name);
        debug!("Git repo remove details: {:?}", &req);

        // Connect to database. Query for the repo
        let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(req.org));

        let db_result = orb_db
            .repo_remove(&req.name)
            .expect("There was a problem removing repo in database");

        debug!("repo remove db result: {:?}", &db_result);

        let org_db = db_result.0;
        let repo_db = db_result.1;
        let secret_db = db_result.2;

        let git_uri_parsed = GitUrl::parse(&repo_db.uri).unwrap();

        // We use auth_data to hold the name of the secret used by repo
        let response = Response::new(GitRepoEntry {
            org: org_db.name,
            git_provider: git_uri_parsed.clone().host.unwrap(),
            name: repo_db.name,
            user: git_uri_parsed.clone().user.unwrap_or_default(),
            uri: format!("{}", git_uri_parsed.trim_auth()),
            secret_type: secret_db.clone().unwrap_or_default().secret_type.into(),
            build: repo_db.build_active_state.into(),
            notify: repo_db.notify_active_state.into(),
            next_build_index: repo_db.next_build_index,
            auth_data: secret_db.unwrap_or_default().name,
            ..Default::default()
        });

        debug!("Response: {:?}", &response);
        Ok(response)
    }

    async fn git_repo_list(
        &self,
        request: Request<GitRepoListRequest>,
    ) -> Result<Response<GitRepoListResponse>, Status> {
        let req = request.into_inner();
        info!("Git repo list request: {:?}", &req);

        // Connect to database. Query for the repo
        let orb_db = postgres::client::OrbitalDBClient::new().set_org(Some(req.org));

        let db_result: Vec<GitRepoEntry> = orb_db
            .repo_list()
            .expect("There was a problem listing repo from database")
            .into_iter()
            .map(|(org_db, repo_db, secret_db)| {
                let git_uri_parsed = GitUrl::parse(&repo_db.uri).unwrap();

                // Convert branch_heads to proto type GitRepoRemoteBranchHeadList
                let mut branch_head_proto_list = GitRepoRemoteBranchHeadList {
                    remote_branch_heads: Vec::new(),
                };
                for (k, v) in repo_db.remote_branch_heads.as_object().unwrap().iter() {
                    let proto_branch = GitRepoRemoteBranchHead {
                        branch: k.to_string(),
                        commit: v.as_str().unwrap().to_string(),
                    };

                    branch_head_proto_list
                        .remote_branch_heads
                        .push(proto_branch);
                }

                GitRepoEntry {
                    org: org_db.name,
                    git_provider: git_uri_parsed.clone().host.unwrap(),
                    name: repo_db.name,
                    user: git_uri_parsed.clone().user.unwrap_or_default(),
                    uri: format!("{}", &git_uri_parsed.trim_auth()),
                    secret_type: secret_db.clone().unwrap_or_default().secret_type.into(),
                    build: repo_db.build_active_state.into(),
                    notify: repo_db.notify_active_state.into(),
                    next_build_index: repo_db.next_build_index,
                    auth_data: secret_db.unwrap_or_default().name,
                    remote_branch_heads: Some(branch_head_proto_list),
                    ..Default::default()
                }
            })
            .collect();

        let git_repos = GitRepoListResponse {
            git_repos: db_result,
        };
        debug!("Response: {:?}", &git_repos);
        Ok(Response::new(git_repos))
    }
}
