use orbital_headers::build_meta::{
    build_service_server::BuildService, BuildLogResponse, BuildMetadata, BuildRecord, BuildStage,
    BuildSummaryRequest, BuildSummaryResponse, BuildTarget,
};

use chrono::{NaiveDateTime, Utc};
use orbital_database::postgres;
use orbital_database::postgres::build_stage::NewBuildStage;
use orbital_database::postgres::build_summary::NewBuildSummary;
use orbital_database::postgres::build_target::NewBuildTarget;
use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
use orbital_headers::orbital_types::{JobState as ProtoJobState, SecretType};
use orbital_headers::secret::{secret_service_client::SecretServiceClient, SecretGetRequest};
use postgres::schema::{JobState, JobTrigger};

use tonic::{Code, Request, Response, Status};

use tokio::sync::mpsc;

//use crate::OrbitalServiceError;
use git_meta::GitCredentials;
use orbital_agent::build_engine;
use orbital_exec_runtime::docker::OrbitalContainerSpec;

use super::state_machine;
use super::{OrbitalApi, ServiceType};

use log::{debug, error, info};

use std::path::Path;
use std::time::Duration;

use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;

use git_meta::git_info;

use serde_json::Value;
use std::str;

// TODO: If this bails anytime before the end, we need to attempt some cleanup
/// Implementation of protobuf derived `BuildService` trait
#[tonic::async_trait]
impl BuildService for OrbitalApi {
    /// Start a build in a container. (Stay focused.)
    ///
    type BuildStartStream = mpsc::Receiver<Result<BuildRecord, Status>>;
    async fn build_start(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<Response<Self::BuildStartStream>, Status> {
        //println!("DEBUG: {:?}", response);

        // Git clone for provider ( uri, branch, commit )
        let unwrapped_request = request.into_inner();

        info!("build request: {:?}", &unwrapped_request.git_repo);
        debug!("build request details: {:?}", &unwrapped_request);

        // TODO: Can we use unbounded_channel() ?
        let (mut client_tx, client_rx) = mpsc::channel(1);

        let (mut build_tx, mut build_rx): (
            mpsc::UnboundedSender<String>,
            mpsc::UnboundedReceiver<String>,
        ) = mpsc::unbounded_channel();

        tokio::spawn(async move {
            let git_clone_dir = Temp::new_dir().expect("Unable to create dir for git clone");

            let mut cur_build = state_machine::BuildContext::new()
                .add_org(unwrapped_request.org.to_string())
                .add_repo_uri(unwrapped_request.clone().remote_uri.to_string())
                .expect("Could not parse repo uri")
                .add_branch(unwrapped_request.branch.to_string())
                .add_hash(unwrapped_request.commit_hash.to_string())
                .add_triggered_by(JobTrigger::Manual)
                .add_working_dir(git_clone_dir.to_path_buf())
                .queue()
                .expect("There was a problem queuing the build");

            if unwrapped_request.config.clone().len() > 0 {
                cur_build = cur_build
                    .clone()
                    .add_build_config_from_string(unwrapped_request.config.clone())
                    .expect("Build config failed to parse");
            }

            'build_loop: loop {
                if (cur_build.clone().state() == state_machine::BuildState::done())
                    | (cur_build.clone().state() == state_machine::BuildState::cancelled())
                    | (cur_build.clone().state() == state_machine::BuildState::fail())
                    | (cur_build.clone().state() == state_machine::BuildState::system_err())
                {
                    debug!("Exiting build loop - {:?}", cur_build.clone().state());
                    break 'build_loop;
                }

                cur_build = cur_build.clone().step(&build_tx).await.unwrap();

                if cur_build.clone().state() == state_machine::BuildState::error() {
                    panic!("State machine error")
                };

                debug!("Trying to listen for output. Not don't block if nothing");
                while let Ok(response) = &build_rx.try_recv() {
                    let mut build_metadata = BuildMetadata {
                        build: Some(unwrapped_request.clone()),
                        job_trigger: cur_build.job_trigger.into(),
                        //job_trigger: JobTrigger::Manual.into(),
                        build_state: ProtoJobState::from(cur_build.clone().state()).into(),
                        ..Default::default()
                    };

                    let mut build_record = BuildRecord {
                        build_metadata: Some(build_metadata.clone()),
                        build_output: Vec::new(),
                    };

                    println!("Stream OUTPUT: {:?}", response.clone().as_str());
                    let mut build_stage_output = BuildStage {
                        ..Default::default()
                    };

                    //println!("PULL OUTPUT: {:?}", response["status"].clone().as_str());
                    //let output = format!("{}\n", response["status"].clone().as_str().unwrap())
                    //    .as_bytes()
                    //    .to_owned();

                    build_stage_output.output = response.as_bytes().to_owned();
                    build_record.build_output.push(build_stage_output);

                    let _ = match client_tx.send(Ok(build_record.clone())).await {
                        Ok(_) => Ok(()),
                        Err(_) => Err(()),
                    };

                    //build_record.build_output.pop(); // Empty out the output buffer
                }

                //    // TODO: Wrap this entire workflow in a check for cancelation
                //    let mut _job_was_canceled = false;

                //    // Parse git info from uri

                //    let mut git_parsed_uri =
                //        git_info::git_remote_url_parse(unwrapped_request.clone().remote_uri.as_ref())
                //            .expect("Could not parse repo uri");

                //    // Maybe leave some global state up here so if we do get canceled, we can modify the db with original start times

                //    let mut build_metadata = BuildMetadata {
                //        build: Some(unwrapped_request.clone()),
                //        job_trigger: JobTrigger::Manual.into(),
                //        build_state: ProtoJobState::Queued.into(),
                //        ..Default::default()
                //    };

                //    let mut build_record = BuildRecord {
                //        build_metadata: Some(build_metadata.clone()),
                //        build_output: Vec::new(),
                //    };

                //    // Add to DB

                //    // Mark the start of build in the database right here
                //    let build_target_record = NewBuildTarget {
                //        git_hash: unwrapped_request.commit_hash.to_string(),
                //        branch: unwrapped_request.branch.to_string(),
                //        user_envs: match &unwrapped_request.user_envs.len() {
                //            0 => None,
                //            _ => Some(unwrapped_request.user_envs.clone()),
                //        },
                //        trigger: unwrapped_request.trigger.clone().into(),
                //        ..Default::default()
                //    };

                //    // Connect to database. Query for the repo
                //    let pg_conn = postgres::client::establish_connection();

                //    // Add build target record in db
                //    debug!("Adding new build target to DB");
                //    let build_target_result = postgres::client::build_target_add(
                //        &pg_conn,
                //        &unwrapped_request.org,
                //        &git_parsed_uri.name,
                //        &build_target_record.git_hash.clone(),
                //        &build_target_record.branch.clone(),
                //        build_target_record.user_envs.clone(),
                //        JobTrigger::Manual.into(),
                //    )
                //    .expect("Build target add failed");

                //    let (org_db, repo_db, build_target_db) = (
                //        build_target_result.0,
                //        build_target_result.1,
                //        build_target_result.2,
                //    );

                //    // TODO: Clean this up by implementing From<BuildTarget> trait
                //    let build_target_current_state = NewBuildTarget {
                //        repo_id: build_target_db.repo_id,
                //        git_hash: build_target_db.git_hash,
                //        branch: build_target_db.branch,
                //        user_envs: build_target_db.user_envs,
                //        queue_time: build_target_db.queue_time,
                //        build_index: build_target_db.build_index,
                //        trigger: build_target_db.trigger,
                //    };

                //    // Add build summary record in db
                //    // Mark build_summary.build_state as queued
                //    let build_summary_current_state = NewBuildSummary {
                //        build_target_id: build_target_db.id,
                //        build_state: postgres::schema::JobState::Queued,
                //        start_time: None,
                //        ..Default::default()
                //    };

                //    // Create a new build summary record
                //    debug!("Adding new build summary to DB");
                //    let _build_summary_result_add = postgres::client::build_summary_add(
                //        &pg_conn,
                //        &org_db.name,
                //        &repo_db.name,
                //        &build_target_current_state.git_hash,
                //        &build_target_current_state.branch,
                //        build_target_current_state.build_index,
                //        build_summary_current_state.clone(),
                //    )
                //    .expect("Unable to create new build summary");

                //    // In the future, this is where the service should return

                //    // BuildMetadata
                //    build_metadata.id = build_target_current_state.build_index;
                //    build_metadata.queue_time = Some(prost_types::Timestamp {
                //        seconds: build_target_db.queue_time.timestamp(),
                //        nanos: build_target_db.queue_time.timestamp_subsec_nanos() as i32,
                //    });

                //    build_record.build_metadata = Some(build_metadata.clone());

                //    // Hack. Reoccurring check Check if build has been canceled
                //    _job_was_canceled = match postgres::client::is_build_canceled(
                //        &pg_conn,
                //        &unwrapped_request.org,
                //        &git_parsed_uri.name,
                //        &unwrapped_request.commit_hash,
                //        &unwrapped_request.branch,
                //        build_target_db.build_index,
                //    ) {
                //        Ok(true) => {
                //            info!("Build has been canceled");
                //            break 'build_loop;
                //        }
                //        Ok(false) => false,
                //        Err(_) => {
                //            //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //            panic!("Error checking for build cancelation")
                //        }
                //    };

                //    build_metadata.build_state = ProtoJobState::Starting.into();

                //    // This is when another thread should start when picking work off queue
                //    // Mark build_summary start time
                //    // Mark build_summary.build_state as starting
                //    let build_summary_current_state = NewBuildSummary {
                //        build_target_id: build_target_db.id,
                //        build_state: postgres::schema::JobState::Starting,
                //        start_time: Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0)),
                //        ..Default::default()
                //    };

                //    // Check if build has been canceled

                //    info!("Updating build state to starting");
                //    let build_summary_result_start = postgres::client::build_summary_update(
                //        &pg_conn,
                //        &org_db.name,
                //        &repo_db.name,
                //        &build_target_current_state.git_hash,
                //        &build_target_current_state.branch,
                //        build_target_current_state.build_index,
                //        build_summary_current_state.clone(),
                //    )
                //    .expect("Unable to update build summary job state to starting");

                //    let (_repo_db, _build_target_db, build_summary_current_state_db) = (
                //        build_summary_result_start.0,
                //        build_summary_result_start.1,
                //        build_summary_result_start.2,
                //    );

                //    build_metadata.start_time = match build_summary_current_state_db.start_time {
                //        Some(t) => Some(prost_types::Timestamp {
                //            seconds: t.timestamp(),
                //            nanos: t.timestamp_subsec_nanos() as i32,
                //        }),
                //        None => None,
                //    };

                //    build_metadata.build_state = ProtoJobState::Starting.into();
                //    build_record.build_metadata = Some(build_metadata.clone());

                //    let _ = match tx.send(Ok(build_record.clone())).await {
                //        Ok(_) => Ok(()),
                //        Err(mpsc::error::SendError(_)) => Err(()),
                //    };

                //    // Hack. Reoccurring check Check if build has been canceled
                //    _job_was_canceled = match postgres::client::is_build_canceled(
                //        &pg_conn,
                //        &unwrapped_request.org,
                //        &git_parsed_uri.name,
                //        &unwrapped_request.commit_hash,
                //        &unwrapped_request.branch,
                //        build_target_db.build_index,
                //    ) {
                //        Ok(true) => {
                //            info!("Build has been canceled");
                //            break 'build_loop;
                //        }
                //        Ok(false) => false,
                //        Err(_) => {
                //            //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //            panic!("Error checking for build cancelation")
                //        }
                //    };

                //    // Retrieve any secrets needed to clone code

                //    debug!("Connecting to the Code service");
                //    let code_client_conn_req = CodeServiceClient::connect(format!(
                //        "http://{}",
                //        super::get_service_uri(ServiceType::Code)
                //    ));

                //    let mut code_client = code_client_conn_req.await.unwrap();

                //    debug!("Building request to Code service for git repo info");

                //    // Request: org/git_provider/name
                //    // e.g.: org_name/github.com/orbitalci/orbital
                //    let request_payload = Request::new(GitRepoGetRequest {
                //        org: unwrapped_request.org.clone().into(),
                //        name: unwrapped_request.clone().git_repo,
                //        uri: unwrapped_request.clone().remote_uri,
                //        ..Default::default()
                //    });

                //    debug!("Payload: {:?}", &request_payload);

                //    debug!("Sending request to Code service for git repo");

                //    let code_service_request = code_client.git_repo_get(request_payload);
                //    let code_service_response = code_service_request.await.unwrap().into_inner();

                //    // Build a GitCredentials struct based on the repo auth type
                //    // Declaring this in case we have an ssh key.
                //    let temp_keypath = Temp::new_file().expect("Unable to create temp file");

                //    // TODO: This is where we're going to get usernames too
                //    // let username, git_creds = ...
                //    let git_creds = match &code_service_response.secret_type.into() {
                //        SecretType::Unspecified => {
                //            // TODO: Call secret service and get a username
                //            info!("No secret needed to clone. Public repo");

                //            GitCredentials::Public
                //        }
                //        SecretType::SshKey => {
                //            info!("SSH key needed to clone");

                //            debug!("Connecting to the Secret service");
                //            let secret_client_conn_req = SecretServiceClient::connect(format!(
                //                "http://{}",
                //                super::get_service_uri(ServiceType::Secret)
                //            ));

                //            let mut secret_client = secret_client_conn_req.await.unwrap();

                //            debug!("Building request to Secret service for git repo ");

                //            // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                //            // Where the secret name is the git repo url
                //            // e.g., "github.com/level11consulting/orbitalci"

                //            let secret_name = format!(
                //                "{}/{}",
                //                &git_parsed_uri.host.clone().expect("No host defined"),
                //                &git_parsed_uri.name,
                //            );

                //            let secret_service_request = Request::new(SecretGetRequest {
                //                org: unwrapped_request.org.clone().into(),
                //                name: secret_name,
                //                secret_type: SecretType::SshKey.into(),
                //                ..Default::default()
                //            });

                //            debug!("Secret request: {:?}", &secret_service_request);

                //            let secret_service_response = secret_client
                //                .secret_get(secret_service_request)
                //                .await
                //                .unwrap()
                //                .into_inner();

                //            debug!("Secret get response: {:?}", &secret_service_response);

                //            // TODO: Deserialize vault data into hashmap.
                //            let vault_response: Value = serde_json::from_str(
                //                str::from_utf8(&secret_service_response.data).unwrap(),
                //            )
                //            .expect("Unable to read json data from Vault");

                //            // Write ssh key to temp file
                //            info!("Writing incoming ssh key to temp file");
                //            let mut file = File::create(temp_keypath.as_path()).unwrap();
                //            let mut _contents = String::new();
                //            let _ = file.write_all(
                //                vault_response["private_key"]
                //                    .as_str()
                //                    .unwrap()
                //                    .to_string()
                //                    .as_bytes(),
                //            );

                //            // TODO: Stop using username from Code service output

                //            // Replace username with the user from the code service
                //            let git_creds = GitCredentials::SshKey {
                //                username: vault_response["username"]
                //                    .clone()
                //                    .as_str()
                //                    .unwrap()
                //                    .to_string(),
                //                public_key: None,
                //                private_key: temp_keypath.as_path(),
                //                passphrase: None,
                //            };

                //            // Add username to git_parsed_uri for cloning
                //            git_parsed_uri.user = Some(
                //                vault_response["username"]
                //                    .clone()
                //                    .as_str()
                //                    .unwrap()
                //                    .to_string(),
                //            );

                //            debug!("Git Creds: {:?}", &git_creds);

                //            git_creds
                //        }
                //        SecretType::BasicAuth => {
                //            info!("Basic Auth creds needed to clone");

                //            debug!("Connecting to the Secret service");
                //            let secret_client_conn_req = SecretServiceClient::connect(format!(
                //                "http://{}",
                //                super::get_service_uri(ServiceType::Secret)
                //            ));
                //            let mut secret_client = secret_client_conn_req.await.unwrap();

                //            debug!("Building request to Secret service for git repo ");

                //            // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                //            // Where the secret name is the git repo url
                //            // e.g., "github.com/level11consulting/orbitalci"

                //            let secret_name = format!(
                //                "{}/{}",
                //                &git_parsed_uri.host.clone().expect("No host defined"),
                //                &git_parsed_uri.name,
                //            );

                //            let secret_service_request = Request::new(SecretGetRequest {
                //                org: unwrapped_request.org.clone().into(),
                //                name: secret_name,
                //                secret_type: SecretType::BasicAuth.into(),
                //                ..Default::default()
                //            });

                //            debug!("Secret request: {:?}", &secret_service_request);

                //            let secret_service_response = secret_client
                //                .secret_get(secret_service_request)
                //                .await
                //                .unwrap()
                //                .into_inner();

                //            debug!("Secret get response: {:?}", &secret_service_response);

                //            // TODO: Deserialize vault data into hashmap.
                //            let vault_response: Value = serde_json::from_str(
                //                str::from_utf8(&secret_service_response.data).unwrap(),
                //            )
                //            .expect("Unable to read json data from Vault");

                //            // Replace username with the user from the code service
                //            let git_creds = GitCredentials::BasicAuth {
                //                username: vault_response["username"].as_str().unwrap().to_string(),
                //                password: vault_response["password"].as_str().unwrap().to_string(),
                //            };

                //            debug!("Git Creds: {:?}", &git_creds);
                //            git_creds
                //        }
                //        _ => panic!(
                //        "We only support public repos, or private repo auth with sshkeys or basic auth"
                //    ),
                //    };

                //    drop(code_service_response);
                //    drop(code_client);

                //    // Hack. Reoccurring check Check if build has been canceled
                //    _job_was_canceled = match postgres::client::is_build_canceled(
                //        &pg_conn,
                //        &unwrapped_request.org,
                //        &git_parsed_uri.name,
                //        &unwrapped_request.commit_hash,
                //        &unwrapped_request.branch,
                //        build_target_db.build_index,
                //    ) {
                //        Ok(true) => {
                //            info!("Build has been canceled");
                //            break 'build_loop;
                //        }
                //        Ok(false) => false,
                //        Err(_) => {
                //            //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //            panic!("Error checking for build cancelation")
                //        }
                //    };

                //    // Clone the code

                //    info!("Cloning code into temp directory");
                //    let git_repo = build_engine::clone_repo(
                //        format!("{}", &git_parsed_uri).as_str(),
                //        &unwrapped_request.branch,
                //        git_creds,
                //    )
                //    .expect("Unable to clone repo");

                //    // build stage end cloning repo.

                //    // Here we parse the newly cloned repo so we can get the commit message
                //    let git_repo_info = git_info::get_git_info_from_path(
                //        git_repo.as_path(),
                //        &Some(unwrapped_request.clone().branch),
                //        &Some(build_target_current_state.clone().git_hash),
                //    )
                //    .unwrap();

                //    // Parse orb.yml from cloned code

                //    let config = match &unwrapped_request.config.len() {
                //        0 => {
                //            debug!("Loading orb.yml from path {:?}", &git_repo.as_path());
                //            build_engine::load_orb_config(Path::new(&format!(
                //                "{}/{}",
                //                &git_repo.as_path().display(),
                //                "orb.yml"
                //            )))
                //            .expect("Unable to load orb.yml")
                //        }
                //        _ => {
                //            debug!("Loading orb.yml from str:\n{:?}", &unwrapped_request.config);
                //            build_engine::load_orb_config_from_str(&unwrapped_request.config)
                //                .expect("Unable to load config from str")
                //        }
                //    };

                //    // Defining internal env vars here
                //    let orb_org_env = format!("ORBITAL_ORG={}", &org_db.name);
                //    let orb_repo_env = format!("ORBITAL_REPOSITORY={}", &repo_db.name);
                //    let orb_build_number_env = format!(
                //        "ORBITAL_BUILD_NUMBER={}",
                //        &build_target_current_state.build_index
                //    );
                //    let orb_commit_env =
                //        format!("ORBITAL_COMMIT={}", &build_target_current_state.git_hash);

                //    let orb_commit_short_env = format!(
                //        "ORBITAL_COMMIT_SHORT={}",
                //        &build_target_current_state.git_hash[0..6]
                //    );
                //    let orb_commit_message = format!("ORBITAL_COMMIT_MSG={}", git_repo_info.message);

                //    // TODO: Need to merge global env vars from config
                //    let orbital_env_vars_vec = vec![
                //        orb_org_env.as_str(),
                //        orb_repo_env.as_str(),
                //        orb_build_number_env.as_str(),
                //        orb_commit_env.as_str(),
                //        orb_commit_short_env.as_str(),
                //        orb_commit_message.as_str(),
                //    ];

                //    // TODO: Use this spec when we can pre-populate the entire build info from config
                //    let build_container_spec = OrbitalContainerSpec {
                //        name: Some(orbital_agent::generate_unique_build_id(
                //            &org_db.name,
                //            &repo_db.name,
                //            &build_target_current_state.git_hash,
                //            &format!("{}", build_target_current_state.build_index),
                //        )),
                //        image: config.image.clone(),
                //        command: Vec::new(), // TODO: Populate this field

                //        // TODO: Inject the dynamic build env vars here
                //        //env_vars: orbital_agent::parse_envs_input(&None),
                //        env_vars: Some(orbital_env_vars_vec),
                //        volumes: orbital_agent::parse_volumes_input(&None),
                //        timeout: Some(Duration::from_secs(crate::DEFAULT_BUILD_TIMEOUT)),
                //    };

                //    // Build Stage start Pulling container

                //    // Hack. Reoccurring check Check if build has been canceled
                //    _job_was_canceled = match postgres::client::is_build_canceled(
                //        &pg_conn,
                //        &unwrapped_request.org,
                //        &git_parsed_uri.name,
                //        &unwrapped_request.commit_hash,
                //        &unwrapped_request.branch,
                //        build_target_db.build_index,
                //    ) {
                //        Ok(true) => {
                //            info!("Build has been canceled");
                //            break 'build_loop;
                //        }
                //        Ok(false) => false,
                //        Err(_) => {
                //            //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //            panic!("Error checking for build cancelation")
                //        }
                //    };

                //    info!(
                //        "Pulling container: {:?}",
                //        build_container_spec.image.clone()
                //    );

                //    //let _ = build_engine::docker_container_pull(&build_container_spec).unwrap();

                //    // I guess here's where I read from the channel?
                //    let mut stream =
                //        build_engine::docker_container_pull_async(build_container_spec.clone())
                //            .await
                //            .unwrap();

                //    while let Some(response) = stream.recv().await {
                //        let mut container_pull_output = BuildStage {
                //            ..Default::default()
                //        };

                //        println!("PULL OUTPUT: {:?}", response["status"].clone().as_str());
                //        let output = format!("{}\n", response["status"].clone().as_str().unwrap())
                //            .as_bytes()
                //            .to_owned();

                //        container_pull_output.output = output;

                //        build_record.build_output.push(container_pull_output);

                //        let _ = match tx.send(Ok(build_record.clone())).await {
                //            Ok(_) => Ok(()),
                //            Err(_) => Err(()),
                //        };

                //        build_record.build_output.pop(); // Empty out the output buffer
                //    }

                //    // Build Stage end Pulling container

                //    // Hack. Reoccurring check Check if build has been canceled
                //    _job_was_canceled = match postgres::client::is_build_canceled(
                //        &pg_conn,
                //        &unwrapped_request.org,
                //        &git_parsed_uri.name,
                //        &unwrapped_request.commit_hash,
                //        &unwrapped_request.branch,
                //        build_target_db.build_index,
                //    ) {
                //        Ok(true) => {
                //            info!("Build has been canceled");
                //            break 'build_loop;
                //        }
                //        Ok(false) => false,
                //        Err(_) => {
                //            //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //            panic!("Error checking for build cancelation")
                //        }
                //    };

                //    // Build Stage start creating container

                //    // Create a new container
                //    info!("Creating container");
                //    let container_id =
                //        build_engine::docker_container_create(&build_container_spec).unwrap();

                //    // Build Stage end creating container

                //    // Hack. Reoccurring check Check if build has been canceled
                //    _job_was_canceled = match postgres::client::is_build_canceled(
                //        &pg_conn,
                //        &unwrapped_request.org,
                //        &git_parsed_uri.name,
                //        &unwrapped_request.commit_hash,
                //        &unwrapped_request.branch,
                //        build_target_db.build_index,
                //    ) {
                //        Ok(true) => {
                //            info!("Build has been canceled");
                //            break 'build_loop;
                //        }
                //        Ok(false) => false,
                //        Err(_) => {
                //            //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //            panic!("Error checking for build cancelation")
                //        }
                //    };

                //    // Build Stage start starting container

                //    // Start a docker container
                //    info!("Start container {:?}", &container_id.as_str());
                //    let _ = build_engine::docker_container_start(container_id.as_str()).unwrap();

                //    // Build Stage end starting container

                //    for (i, config_stage) in config.stages.iter().enumerate() {
                //        // Hack. Reoccurring check Check if build has been canceled
                //        _job_was_canceled = match postgres::client::is_build_canceled(
                //            &pg_conn,
                //            &unwrapped_request.org,
                //            &git_parsed_uri.name,
                //            &unwrapped_request.commit_hash,
                //            &unwrapped_request.branch,
                //            build_target_db.build_index,
                //        ) {
                //            Ok(true) => {
                //                info!("Build has been canceled");
                //                break 'build_loop;
                //            }
                //            Ok(false) => false,
                //            Err(_) => {
                //                //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //                panic!("Error checking for build cancelation")
                //            }
                //        };

                //        debug!(
                //            "Starting stage: {:?}",
                //            &config_stage.name.clone().unwrap_or(format!("Stage#{}", i))
                //        );

                //        // FIXME: Loop over config.command and run the docker_container_exec, for timestamping build_stage
                //        // TODO: Mark build summary.build_state as running
                //        let build_summary_current_state = NewBuildSummary {
                //            build_target_id: build_summary_current_state_db.build_target_id,
                //            start_time: build_summary_current_state_db.start_time,
                //            build_state: postgres::schema::JobState::Running,
                //            ..Default::default()
                //        };

                //        // Build Stage running config_stage.name

                //        info!("Updating build state to running");
                //        let build_summary_result_running = postgres::client::build_summary_update(
                //            &pg_conn,
                //            &org_db.name,
                //            &repo_db.name,
                //            &build_target_current_state.git_hash,
                //            &build_target_current_state.branch,
                //            build_target_current_state.build_index,
                //            build_summary_current_state.clone(),
                //        )
                //        .expect("Unable to update build summary job state to running");

                //        let (_repo_db, _build_target_db, build_summary_current_state_db) = (
                //            build_summary_result_running.0,
                //            build_summary_result_running.1,
                //            build_summary_result_running.2,
                //        );

                //        build_metadata.build_state = ProtoJobState::Running.into();
                //        build_record.build_metadata = Some(build_metadata.clone());

                //        let _ = match tx.send(Ok(build_record.clone())).await {
                //            Ok(_) => Ok(()),
                //            Err(mpsc::error::SendError(_)) => Err(()),
                //        };

                //        // TODO: Mark build stage start time
                //        //let build_stage_current = NewBuildStage {
                //        //    build_summary_id: build_summary_current_state_db.id,
                //        //    stage_name: Some(config_stage.name.clone().unwrap_or(format!("Stage#{}", i))),
                //        //    ..Default::default()
                //        //};

                //        info!("Adding build stage");
                //        let build_stage_start = postgres::client::build_stage_add(
                //            &pg_conn,
                //            &org_db.name,
                //            &repo_db.name,
                //            &build_target_current_state.git_hash,
                //            &build_target_current_state.branch,
                //            build_target_current_state.build_index,
                //            build_summary_current_state_db.id,
                //            NewBuildStage {
                //                build_summary_id: build_summary_current_state_db.id,
                //                stage_name: Some(
                //                    config_stage.name.clone().unwrap_or(format!("Stage#{}", i)),
                //                ),
                //                build_host: Some("Hardcoded hostname Fixme".to_string()),
                //                ..Default::default()
                //            },
                //        )
                //        .expect("Unable to add new build stage in db");

                //        let (_build_target_db, _build_summary_db, build_stage_db) = (
                //            build_stage_start.0,
                //            build_stage_start.1,
                //            build_stage_start.2,
                //        );

                //        // TODO: Make sure tests try to exec w/o starting the container
                //        // Exec into the new container
                //        debug!("Sending commands into container");

                //        //let build_stage_logs = build_engine::docker_container_exec(
                //        //    container_id.as_str(),
                //        //    config_stage.command.clone(),
                //        //)
                //        //.unwrap();

                //        let mut build_stage_logs = String::new();

                //        // TODO: Probably need to break up command so we can granularly check for cancelation

                //        let mut stream = build_engine::docker_container_exec_async(
                //            container_id.clone(),
                //            config_stage.command.clone(),
                //        )
                //        .await
                //        .unwrap();

                //        while let Some(response) = stream.recv().await {
                //            let mut container_exec_output = BuildStage {
                //                ..Default::default()
                //            };

                //            println!("EXEC OUTPUT: {:?}", response.clone().as_str());

                //            // Adding newlines

                //            let output = response.clone().as_bytes().to_owned();

                //            container_exec_output.output = output;

                //            build_stage_logs.push_str(response.clone().as_str());

                //            build_record.build_output.push(container_exec_output);

                //            let _ = match tx.send(Ok(build_record.clone())).await {
                //                Ok(_) => Ok(()),
                //                Err(mpsc::error::SendError(_)) => Err(()),
                //            };

                //            build_record.build_output.pop(); // Empty out the output buffer
                //        }

                //        // Build Stage finishing config_stage.name

                //        // Hack. Reoccurring check Check if build has been canceled
                //        _job_was_canceled = match postgres::client::is_build_canceled(
                //            &pg_conn,
                //            &unwrapped_request.org,
                //            &git_parsed_uri.name,
                //            &unwrapped_request.commit_hash,
                //            &unwrapped_request.branch,
                //            build_target_db.build_index,
                //        ) {
                //            Ok(true) => {
                //                info!("Build has been canceled");
                //                break 'build_loop;
                //            }
                //            Ok(false) => false,
                //            Err(_) => {
                //                //return Err(Status::new(Code::Internal, "Error checking for build cancelation"))
                //                panic!("Error checking for build cancelation")
                //            }
                //        };

                //        // Mark build_summary.build_state as finishing
                //        let build_summary_current_state = NewBuildSummary {
                //            build_target_id: build_summary_current_state_db.build_target_id,
                //            start_time: build_summary_current_state_db.start_time,
                //            build_state: postgres::schema::JobState::Finishing,
                //            ..Default::default()
                //        };

                //        info!("Updating build state to finishing");
                //        let build_summary_result_finishing = postgres::client::build_summary_update(
                //            &pg_conn,
                //            &org_db.name,
                //            &repo_db.name,
                //            &build_target_current_state.git_hash,
                //            &build_target_current_state.branch,
                //            build_target_current_state.build_index,
                //            build_summary_current_state.clone(),
                //        )
                //        .expect("Unable to update build summary job state to running");

                //        let (_repo_db, _build_target_db, _build_summary_current_state_db) = (
                //            build_summary_result_finishing.0,
                //            build_summary_result_finishing.1,
                //            build_summary_result_finishing.2,
                //        );

                //        // Mark build stage end time and save stage output
                //        info!("Marking end of build stage");
                //        let _build_stage_end = postgres::client::build_stage_update(
                //            &pg_conn,
                //            &org_db.name,
                //            &repo_db.name,
                //            &build_target_current_state.git_hash,
                //            &build_target_current_state.branch,
                //            build_target_current_state.build_index,
                //            build_summary_current_state_db.id,
                //            build_stage_db.id,
                //            NewBuildStage {
                //                build_summary_id: build_summary_current_state_db.id,
                //                stage_name: Some(
                //                    config_stage.name.clone().unwrap_or(format!("Stage#{}", i)),
                //                ),
                //                start_time: build_stage_db.start_time,
                //                end_time: Some(NaiveDateTime::from_timestamp(
                //                    Utc::now().timestamp(),
                //                    0,
                //                )),
                //                build_host: build_stage_db.build_host,
                //                output: Some(build_stage_logs),
                //                ..Default::default()
                //            },
                //        );
                //        // END Looping over stages
                //    }

                //    info!("Stopping the container");
                //    let _ = build_engine::docker_container_stop(container_id.as_str()).unwrap();

                //    // Mark build_summary end time
                //    // Mark build_summary.build_statue as done
                //    let build_summary_current_state = NewBuildSummary {
                //        build_target_id: build_summary_current_state_db.build_target_id,
                //        start_time: build_summary_current_state_db.start_time,
                //        end_time: Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0)),
                //        build_state: postgres::schema::JobState::Done,
                //        ..Default::default()
                //    };

                //    info!("Updating build state to done");
                //    let build_summary_result_done = postgres::client::build_summary_update(
                //        &pg_conn,
                //        &org_db.name,
                //        &repo_db.name,
                //        &build_target_current_state.git_hash,
                //        &build_target_current_state.branch,
                //        build_target_current_state.build_index,
                //        build_summary_current_state.clone(),
                //    )
                //    .expect("Unable to update build summary job state to done");

                //    let (_repo_db, _build_target_db, build_summary_current_state_db) = (
                //        build_summary_result_done.0,
                //        build_summary_result_done.1,
                //        build_summary_result_done.2,
                //    );

                //    build_metadata.end_time = match build_summary_current_state_db.end_time {
                //        Some(t) => Some(prost_types::Timestamp {
                //            seconds: t.timestamp(),
                //            nanos: t.timestamp_subsec_nanos() as i32,
                //        }),
                //        None => None,
                //    };

                //    build_metadata.build_state = ProtoJobState::Done.into();
                //    build_record.build_metadata = Some(build_metadata.clone());

                //    let _ = match tx.send(Ok(build_record.clone())).await {
                //        Ok(_) => Ok(()),
                //        Err(mpsc::error::SendError(_)) => Err(()),
                //    };
                //    break 'build_loop;
            }
        });

        Ok(Response::new(client_rx))
    }

    async fn build_stop(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        let unwrapped_request = request.into_inner();

        let pg_conn = postgres::client::establish_connection();

        // Resolve the build number to latest if build number is 0
        let build_id = match unwrapped_request.id {
            0 => {
                if let Ok((_, repo, _)) = postgres::client::repo_get(
                    &pg_conn,
                    &unwrapped_request.org,
                    &unwrapped_request.git_repo,
                ) {
                    repo.next_build_index - 1
                } else {
                    panic!("No build id provided. Failed to query DB for latest build id")
                }
            }
            _ => unwrapped_request.id,
        };

        // Determine if build is cancelable
        match postgres::client::build_summary_get(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.git_repo,
            &unwrapped_request.commit_hash,
            &unwrapped_request.branch,
            build_id,
        ) {
            Ok((repo, build_target, Some(summary))) => match summary.build_state {
                JobState::Queued => {
                    info!("Stop build before it even gets started");

                    // Probably change the build job state to canceled
                    let mut new_canceled_summary = summary.clone();
                    new_canceled_summary.build_state = JobState::Canceled;

                    info!("Updating build state to canceled");
                    let _build_summary_result_canceled = postgres::client::build_summary_update(
                        &pg_conn,
                        &unwrapped_request.org,
                        &repo.name,
                        &build_target.git_hash,
                        &build_target.branch,
                        build_target.build_index,
                        NewBuildSummary {
                            build_target_id: summary.build_target_id,
                            start_time: summary.start_time,
                            end_time: Some(NaiveDateTime::from_timestamp(
                                Utc::now().timestamp(),
                                0,
                            )),
                            build_state: postgres::schema::JobState::Canceled,
                            ..Default::default()
                        },
                    )
                    .expect("Unable to update build summary job state to canceled");

                    Ok(Response::new(BuildMetadata {
                        ..Default::default()
                    }))
                }
                JobState::Starting | JobState::Running => {
                    // Send build cancelation signal
                    let container_name = orbital_agent::generate_unique_build_id(
                        &unwrapped_request.org,
                        &unwrapped_request.git_repo,
                        &unwrapped_request.commit_hash,
                        &format!("{}", build_id),
                    );

                    info!("Send a cancel signal for container: {}", &container_name);

                    // Probably change the build job state to canceled
                    let _ = build_engine::docker_container_stop(&container_name)
                        .expect("Sending Docker container stop failed");

                    // Update summary.build_state to JobState::Canceled
                    let mut new_canceled_summary = summary.clone();
                    new_canceled_summary.build_state = JobState::Canceled;

                    info!("Updating build state to canceled");
                    let _build_summary_result_canceled = postgres::client::build_summary_update(
                        &pg_conn,
                        &unwrapped_request.org,
                        &repo.name,
                        &build_target.git_hash,
                        &build_target.branch,
                        build_target.build_index,
                        NewBuildSummary {
                            build_target_id: summary.build_target_id,
                            start_time: summary.start_time,
                            end_time: Some(NaiveDateTime::from_timestamp(
                                Utc::now().timestamp(),
                                0,
                            )),
                            build_state: postgres::schema::JobState::Canceled,
                            ..Default::default()
                        },
                    )
                    .expect("Unable to update build summary job state to canceled");

                    Ok(Response::new(BuildMetadata {
                        ..Default::default()
                    }))
                }
                _ => {
                    println!("Build is not cancelable");
                    Err(Status::new(Code::Aborted, "Build not cancelable"))
                }
            },
            Ok((_, _, None)) => {
                // Build hasn't been queued yet
                error!("Build is not yet queued, and couldn't be canceled. This is a bug.");
                Err(Status::new(
                    Code::FailedPrecondition,
                    "FIXME: Build has not been queued yet but we can't signal a cancel",
                ))
            }
            Err(_) => {
                error!("Build was not found");
                Err(Status::new(Code::NotFound, "Build was not found"))
            }
        }
    }

    //type BuildLogsStream =
    //    Pin<Box<dyn Stream<Item = Result<BuildLogResponse, Status>> + Send + Sync + 'static>>;

    type BuildLogsStream = mpsc::Receiver<Result<BuildLogResponse, Status>>;

    async fn build_logs(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<tonic::Response<Self::BuildLogsStream>, tonic::Status> {
        let unwrapped_request = request.into_inner();

        // Get repo id from BuildTarget
        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        //let build_stage_query = postgres::client::build_logs_get(
        //    &pg_conn,
        //    &unwrapped_request.org,
        //    &unwrapped_request.git_repo,
        //    &unwrapped_request.commit_hash,
        //    &unwrapped_request.branch,
        //    {
        //        match &unwrapped_request.id {
        //            0 => None,
        //            _ => Some(unwrapped_request.id.clone()),
        //        }
        //    },
        //)
        //.expect("No build stages found");

        // Resolve the build number to latest if build number is 0
        let build_id = match unwrapped_request.id {
            0 => {
                if let Ok((_, repo, _)) = postgres::client::repo_get(
                    &pg_conn,
                    &unwrapped_request.org,
                    &unwrapped_request.git_repo,
                ) {
                    repo.next_build_index - 1
                } else {
                    panic!("No build id provided. Failed to query DB for latest build id")
                }
            }
            _ => unwrapped_request.id,
        };

        let (_repo, _build_target, build_summary_opt) = postgres::client::build_summary_get(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.git_repo,
            &unwrapped_request.commit_hash,
            &unwrapped_request.branch,
            build_id,
        )
        .unwrap();

        drop(pg_conn);

        let (mut tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            match build_summary_opt {
                Some(summary) => {
                    match summary.build_state {
                        JobState::Queued | JobState::Running => {
                            let container_name = orbital_agent::generate_unique_build_id(
                                &unwrapped_request.org,
                                &unwrapped_request.git_repo,
                                &unwrapped_request.commit_hash,
                                &format!("{}", build_id),
                            );

                            let mut stream =
                                build_engine::docker_container_logs(container_name.clone())
                                    .await
                                    .unwrap();

                            while let Some(response) = stream.recv().await {
                                let mut container_logs = BuildStage {
                                    ..Default::default()
                                };

                                println!("LOGS OUTPUT: {:?}", response.clone().as_str());

                                // Adding newlines

                                let output = response.clone().as_bytes().to_owned();
                                container_logs.output = output;

                                let build_record = BuildRecord {
                                    build_metadata: None,
                                    build_output: vec![container_logs],
                                };

                                //
                                let build_log_response = BuildLogResponse {
                                    id: build_id,
                                    records: vec![build_record],
                                };

                                let _ = match tx.send(Ok(build_log_response)).await {
                                    Ok(_) => Ok(()),
                                    Err(mpsc::error::SendError(_)) => Err(()),
                                };
                            }
                        }

                        _ => {
                            let pg_conn = postgres::client::establish_connection();
                            let build_stage_query = postgres::client::build_logs_get(
                                &pg_conn,
                                &unwrapped_request.org,
                                &unwrapped_request.git_repo,
                                &unwrapped_request.commit_hash,
                                &unwrapped_request.branch,
                                Some(build_id),
                            )
                            .expect("No build stages found");

                            let mut build_stage_list: Vec<orbital_headers::build_meta::BuildStage> =
                                Vec::new();
                            for (_target, _summary, stage) in build_stage_query {
                                build_stage_list.push(stage.into());
                            }

                            let build_record = BuildRecord {
                                build_metadata: None,
                                build_output: build_stage_list,
                            };

                            //
                            let build_log_response = BuildLogResponse {
                                id: build_record.build_output[0].build_id,
                                records: vec![build_record],
                            };

                            let _ = match tx.send(Ok(build_log_response)).await {
                                Ok(_) => Ok(()),
                                Err(mpsc::error::SendError(_)) => Err(()),
                            };
                        }
                    }
                }
                None => (),
            }
        });

        Ok(Response::new(rx))
    }

    async fn build_remove(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        unimplemented!();
    }

    async fn build_summary(
        &self,
        request: Request<BuildSummaryRequest>,
    ) -> Result<Response<BuildSummaryResponse>, Status> {
        let unwrapped_request = request.into_inner();
        let build_info = &unwrapped_request
            .build
            .clone()
            .expect("No build info provided");

        debug!("Received request: {:?}", &unwrapped_request);

        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        let build_summary_db = postgres::client::build_summary_list(
            &pg_conn,
            &build_info.org,
            &build_info.git_repo,
            unwrapped_request.limit,
        )
        .expect("No summary returned");

        debug!("Summary: {:?}", &build_summary_db);

        let metadata_proto: Vec<BuildMetadata> = build_summary_db
            .into_iter()
            .map(|(repo, target, summary)| BuildMetadata {
                id: summary.id,
                build: Some(BuildTarget {
                    org: build_info.org.clone(),
                    git_repo: repo.name,
                    remote_uri: repo.uri,
                    branch: target.branch,
                    commit_hash: target.git_hash,
                    user_envs: match target.user_envs {
                        Some(e) => e,
                        None => "".to_string(),
                    },
                    id: target.build_index,
                    trigger: target.trigger.into(),
                    config: "".to_string(),
                }),
                job_trigger: target.trigger.into(),
                queue_time: Some(prost_types::Timestamp {
                    seconds: target.queue_time.timestamp(),
                    nanos: target.queue_time.timestamp_subsec_nanos() as i32,
                }),
                start_time: match summary.start_time {
                    Some(start_time) => Some(prost_types::Timestamp {
                        seconds: start_time.timestamp(),
                        nanos: start_time.timestamp_subsec_nanos() as i32,
                    }),
                    None => None,
                },
                end_time: match summary.end_time {
                    Some(end_time) => Some(prost_types::Timestamp {
                        seconds: end_time.timestamp(),
                        nanos: end_time.timestamp_subsec_nanos() as i32,
                    }),
                    None => None,
                },
                build_state: summary.build_state.into(),
            })
            .collect();

        Ok(Response::new(BuildSummaryResponse {
            summaries: metadata_proto,
        }))
    }
}
