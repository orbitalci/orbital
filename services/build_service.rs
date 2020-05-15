use orbital_headers::build_meta::{
    build_service_server::BuildService,
    BuildLogResponse,
    BuildMetadata,
    BuildRecord, //BuildStage,
    BuildSummaryRequest,
    BuildSummaryResponse,
    BuildTarget,
};

use chrono::{NaiveDateTime, Utc};
use orbital_database::postgres;
use orbital_database::postgres::build_stage::NewBuildStage;
use orbital_database::postgres::build_summary::NewBuildSummary;
use orbital_database::postgres::build_target::NewBuildTarget;
use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
use orbital_headers::orbital_types::{JobState, SecretType};
use orbital_headers::secret::{secret_service_client::SecretServiceClient, SecretGetRequest};
use postgres::schema::JobTrigger;

use tonic::{Request, Response, Status};

use tokio::sync::mpsc;

use crate::OrbitalServiceError;
use agent_runtime::build_engine;
use git_meta::GitCredentials;

use super::{OrbitalApi, ServiceType};

use log::{debug, info};

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
    async fn build_start(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        //println!("DEBUG: {:?}", response);

        // Git clone for provider ( uri, branch, commit )
        let unwrapped_request = request.into_inner();
        info!("build request: {:?}", &unwrapped_request.git_repo);
        debug!("build request details: {:?}", &unwrapped_request);

        let mut git_parsed_uri =
            git_info::git_remote_url_parse(unwrapped_request.clone().remote_uri.as_ref())
                .expect("Could not parse repo uri");

        // Placeholder for things we need to check for eventually when we support more things
        // Org - We want the user to select their org, and cache it locally
        // Is Repo in the system? Halt if not
        // Git provider - to select APIs - a property of the repo. This is more useful if self-hosting git
        // Determine the type of secret
        // Connect to SecretService: Get the creds for the repo
        // create write a temporary file to hold the key, so we can pass it to the git clone

        // Connect to SecretService: Collect all env vars

        // Ocelot current takes all of the secrets from an org and throws it into the container.
        // We should probably follow suit until we know better

        // Connect to the code service to deref the repo to the secret it uses

        debug!("Connecting to the Code service");
        let code_client_conn_req = CodeServiceClient::connect(format!(
            "http://{}",
            super::get_service_uri(ServiceType::Code)
        ));
        let mut code_client = match code_client_conn_req.await {
            Ok(connection_handler) => connection_handler,
            Err(_e) => {
                return Err(OrbitalServiceError::new("Unable to connect to Code service").into())
            }
        };

        debug!("Building request to Code service for git repo info");

        // Request: org/git_provider/name
        // e.g.: org_name/github.com/orbitalci/orbital
        let request_payload = Request::new(GitRepoGetRequest {
            org: unwrapped_request.org.clone().into(),
            name: unwrapped_request.clone().git_repo,
            uri: unwrapped_request.clone().remote_uri,
            ..Default::default()
        });

        debug!("Payload: {:?}", &request_payload);

        debug!("Sending request to Code service for git repo");
        let code_service_request = code_client.git_repo_get(request_payload);
        let code_service_response = match code_service_request.await {
            Ok(r) => {
                debug!("Git repo get response: {:?}", &r);
                r.into_inner()
            }
            Err(_e) => {
                return Err(OrbitalServiceError::new("There was an error getting git repo").into())
            }
        };

        // Build a GitCredentials struct based on the repo auth type
        // Declaring this in case we have an ssh key.
        let temp_keypath = Temp::new_file().expect("Unable to create temp file");

        // TODO: This is where we're going to get usernames too
        // let username, git_creds = ...
        let git_creds = match &code_service_response.secret_type.into() {
            SecretType::Unspecified => {
                // TODO: Call secret service and get a username
                info!("No secret needed to clone. Public repo");

                GitCredentials::Public
            }
            SecretType::SshKey => {
                info!("SSH key needed to clone");

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

                debug!("Building request to Secret service for git repo ");

                // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                // Where the secret name is the git repo url
                // e.g., "github.com/level11consulting/orbitalci"

                let secret_name = format!(
                    "{}/{}",
                    &git_parsed_uri.host.clone().expect("No host defined"),
                    &git_parsed_uri.name,
                );

                let secret_service_request = Request::new(SecretGetRequest {
                    org: unwrapped_request.org.clone().into(),
                    name: secret_name,
                    secret_type: SecretType::SshKey.into(),
                    ..Default::default()
                });

                debug!("Secret request: {:?}", &secret_service_request);

                let secret_service_response =
                    match secret_client.secret_get(secret_service_request).await {
                        Ok(r) => r.into_inner(),
                        Err(_e) => {
                            return Err(OrbitalServiceError::new(
                                "There was an error getting git repo",
                            )
                            .into())
                        }
                    };

                debug!("Secret get response: {:?}", &secret_service_response);

                // TODO: Deserialize vault data into hashmap.
                let vault_response: Value =
                    serde_json::from_str(str::from_utf8(&secret_service_response.data).unwrap())
                        .expect("Unable to read json data from Vault");

                // Write ssh key to temp file
                info!("Writing incoming ssh key to temp file");
                let mut file = File::create(temp_keypath.as_path())?;
                let mut _contents = String::new();
                let _ = file.write_all(
                    vault_response["private_key"]
                        .as_str()
                        .unwrap()
                        .to_string()
                        .as_bytes(),
                );

                // TODO: Stop using username from Code service output

                // Replace username with the user from the code service
                let git_creds = GitCredentials::SshKey {
                    username: vault_response["username"]
                        .clone()
                        .as_str()
                        .unwrap()
                        .to_string(),
                    public_key: None,
                    private_key: temp_keypath.as_path(),
                    passphrase: None,
                };

                // Add username to git_parsed_uri for cloning
                git_parsed_uri.user = Some(
                    vault_response["username"]
                        .clone()
                        .as_str()
                        .unwrap()
                        .to_string(),
                );

                debug!("Git Creds: {:?}", &git_creds);

                git_creds
            }
            SecretType::BasicAuth => {
                info!("Basic Auth creds needed to clone");

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

                debug!("Building request to Secret service for git repo ");

                // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                // Where the secret name is the git repo url
                // e.g., "github.com/level11consulting/orbitalci"

                let secret_name = format!(
                    "{}/{}",
                    &git_parsed_uri.host.clone().expect("No host defined"),
                    &git_parsed_uri.name,
                );

                let secret_service_request = Request::new(SecretGetRequest {
                    org: unwrapped_request.org.clone().into(),
                    name: secret_name,
                    secret_type: SecretType::BasicAuth.into(),
                    ..Default::default()
                });

                debug!("Secret request: {:?}", &secret_service_request);

                let secret_service_response =
                    match secret_client.secret_get(secret_service_request).await {
                        Ok(r) => r.into_inner(),
                        Err(_e) => {
                            return Err(OrbitalServiceError::new(
                                "There was an error getting git repo",
                            )
                            .into())
                        }
                    };

                debug!("Secret get response: {:?}", &secret_service_response);

                // TODO: Deserialize vault data into hashmap.
                let vault_response: Value =
                    serde_json::from_str(str::from_utf8(&secret_service_response.data).unwrap())
                        .expect("Unable to read json data from Vault");

                // Replace username with the user from the code service
                let git_creds = GitCredentials::BasicAuth {
                    username: vault_response["username"].as_str().unwrap().to_string(),
                    password: vault_response["password"].as_str().unwrap().to_string(),
                };

                debug!("Git Creds: {:?}", &git_creds);
                git_creds
            }
            _ => panic!(
                "We only support public repos, or private repo auth with sshkeys or basic auth"
            ),
        };

        // Mark the start of build in the database right here
        let build_target_record = NewBuildTarget {
            //name: git_parsed_uri.name.to_string(),
            git_hash: unwrapped_request.commit_hash.to_string(),
            branch: unwrapped_request.branch.to_string(),
            user_envs: match &unwrapped_request.user_envs.len() {
                0 => None,
                _ => Some(unwrapped_request.user_envs.clone()),
            },
            trigger: unwrapped_request.trigger.clone().into(),

            ..Default::default()
        };

        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        // Add build target record in db
        debug!("Adding new build target to DB");
        let build_target_result = postgres::client::build_target_add(
            &pg_conn,
            &unwrapped_request.org,
            &git_parsed_uri.name,
            &build_target_record.git_hash.clone(),
            &build_target_record.branch.clone(),
            build_target_record.user_envs.clone(),
            JobTrigger::Manual.into(),
        )
        .expect("Build target add failed");

        let (org_db, repo_db, build_target_db) = (
            build_target_result.0,
            build_target_result.1,
            build_target_result.2,
        );

        // TODO: Clean this up by implementing From<BuildTarget> trait
        let build_target_current_state = NewBuildTarget {
            repo_id: build_target_db.repo_id,
            git_hash: build_target_db.git_hash,
            branch: build_target_db.branch,
            user_envs: build_target_db.user_envs,
            queue_time: build_target_db.queue_time,
            build_index: build_target_db.build_index,
            trigger: build_target_db.trigger,
        };

        // Add build summary record in db
        // Mark build_summary.build_state as queued
        let build_summary_current_state = NewBuildSummary {
            build_target_id: build_target_db.id,
            build_state: postgres::schema::JobState::Queued,
            start_time: None,
            ..Default::default()
        };

        // Create a new build summary record
        debug!("Adding new build summary to DB");
        let _build_summary_result_add = postgres::client::build_summary_add(
            &pg_conn,
            &org_db.name,
            &repo_db.name,
            &build_target_current_state.git_hash,
            &build_target_current_state.branch,
            build_target_current_state.build_index,
            build_summary_current_state.clone(),
        )
        .expect("Unable to create new build summary");

        // In the future, this is where the service should return

        // This is when another thread should start when picking work off queue
        // Mark build_summary start time
        // Mark build_summary.build_state as starting
        let build_summary_current_state = NewBuildSummary {
            build_target_id: build_target_db.id,
            build_state: postgres::schema::JobState::Starting,
            start_time: Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0)),
            ..Default::default()
        };

        info!("Updating build state to starting");
        let build_summary_result_start = postgres::client::build_summary_update(
            &pg_conn,
            &org_db.name,
            &repo_db.name,
            &build_target_current_state.git_hash,
            &build_target_current_state.branch,
            build_target_current_state.build_index,
            build_summary_current_state.clone(),
        )
        .expect("Unable to update build summary job state to starting");

        let (_repo_db, _build_target_db, build_summary_current_state_db) = (
            build_summary_result_start.0,
            build_summary_result_start.1,
            build_summary_result_start.2,
        );

        // TODO: Replace expect with match, so we can update db in case of failures
        info!("Cloning code into temp directory");
        let git_repo = build_engine::clone_repo(
            format!("{}", &git_parsed_uri).as_str(),
            &unwrapped_request.branch,
            git_creds,
        )
        .expect("Unable to clone repo");

        let config = match &unwrapped_request.config.len() {
            0 => {
                debug!("Loading orb.yml from path {:?}", &git_repo.as_path());
                build_engine::load_orb_config(Path::new(&format!(
                    "{}/{}",
                    &git_repo.as_path().display(),
                    "orb.yml"
                )))
                .expect("Unable to load orb.yml")
            }
            _ => {
                debug!("Loading orb.yml from str:\n{:?}", &unwrapped_request.config);
                build_engine::load_orb_config_from_str(&unwrapped_request.config)
                    .expect("Unable to load config from str")
            }
        };

        info!("Pulling container: {:?}", config.image.clone());
        match build_engine::docker_container_pull(config.image.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // TODO: Inject the dynamic build env vars here
        let envs_vec = agent_runtime::parse_envs_input(&None);
        let vols_vec = agent_runtime::parse_volumes_input(&None);

        // Create a new container
        info!("Creating container");
        let container_id = match build_engine::docker_container_create(
            config.image.as_str(),
            envs_vec,
            vols_vec,
            Duration::from_secs(crate::DEFAULT_BUILD_TIMEOUT),
        ) {
            Ok(id) => id,
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // Start a docker container
        info!("Start container");
        match build_engine::docker_container_start(container_id.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        for (i, config_stage) in config.stages.iter().enumerate() {
            debug!(
                "Starting stage: {:?}",
                &config_stage.name.clone().unwrap_or(format!("Stage#{}", i))
            );

            // FIXME: Loop over config.command and run the docker_container_exec, for timestamping build_stage
            // TODO: Mark build summary.build_state as running
            let build_summary_current_state = NewBuildSummary {
                build_target_id: build_summary_current_state_db.build_target_id,
                start_time: build_summary_current_state_db.start_time,
                build_state: postgres::schema::JobState::Running,
                ..Default::default()
            };

            info!("Updating build state to running");
            let build_summary_result_running = postgres::client::build_summary_update(
                &pg_conn,
                &org_db.name,
                &repo_db.name,
                &build_target_current_state.git_hash,
                &build_target_current_state.branch,
                build_target_current_state.build_index,
                build_summary_current_state.clone(),
            )
            .expect("Unable to update build summary job state to running");

            let (_repo_db, _build_target_db, build_summary_current_state_db) = (
                build_summary_result_running.0,
                build_summary_result_running.1,
                build_summary_result_running.2,
            );

            // TODO: Mark build stage start time
            //let build_stage_current = NewBuildStage {
            //    build_summary_id: build_summary_current_state_db.id,
            //    stage_name: Some(config_stage.name.clone().unwrap_or(format!("Stage#{}", i))),
            //    ..Default::default()
            //};

            info!("Adding build stage");
            let build_stage_start = postgres::client::build_stage_add(
                &pg_conn,
                &org_db.name,
                &repo_db.name,
                &build_target_current_state.git_hash,
                &build_target_current_state.branch,
                build_target_current_state.build_index,
                build_summary_current_state_db.id,
                NewBuildStage {
                    build_summary_id: build_summary_current_state_db.id,
                    stage_name: Some(config_stage.name.clone().unwrap_or(format!("Stage#{}", i))),
                    build_host: Some("Hardcoded hostname Fixme".to_string()),
                    ..Default::default()
                },
            )
            .expect("Unable to add new build stage in db");

            let (_build_target_db, _build_summary_db, build_stage_db) = (
                build_stage_start.0,
                build_stage_start.1,
                build_stage_start.2,
            );

            // TODO: Make sure tests try to exec w/o starting the container
            // Exec into the new container
            debug!("Sending commands into container");
            let build_stage_logs = match build_engine::docker_container_exec(
                container_id.as_str(),
                config_stage.command.clone(),
            ) {
                Ok(ok) => ok, // The successful result doesn't matter
                Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
            };

            // Mark build_summary.build_state as finishing
            let build_summary_current_state = NewBuildSummary {
                build_target_id: build_summary_current_state_db.build_target_id,
                start_time: build_summary_current_state_db.start_time,
                build_state: postgres::schema::JobState::Finishing,
                ..Default::default()
            };

            info!("Updating build state to finishing");
            let build_summary_result_finishing = postgres::client::build_summary_update(
                &pg_conn,
                &org_db.name,
                &repo_db.name,
                &build_target_current_state.git_hash,
                &build_target_current_state.branch,
                build_target_current_state.build_index,
                build_summary_current_state.clone(),
            )
            .expect("Unable to update build summary job state to running");

            let (_repo_db, _build_target_db, _build_summary_current_state_db) = (
                build_summary_result_finishing.0,
                build_summary_result_finishing.1,
                build_summary_result_finishing.2,
            );

            // Mark build stage end time and save stage output
            info!("Marking end of build stage");
            let _build_stage_end = postgres::client::build_stage_update(
                &pg_conn,
                &org_db.name,
                &repo_db.name,
                &build_target_current_state.git_hash,
                &build_target_current_state.branch,
                build_target_current_state.build_index,
                build_summary_current_state_db.id,
                build_stage_db.id,
                NewBuildStage {
                    build_summary_id: build_summary_current_state_db.id,
                    stage_name: Some(config_stage.name.clone().unwrap_or(format!("Stage#{}", i))),
                    start_time: build_stage_db.start_time,
                    end_time: Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0)),
                    build_host: build_stage_db.build_host,
                    output: Some(build_stage_logs),
                    ..Default::default()
                },
            );
            // END Looping over stages
        }

        info!("Stopping the container");
        match build_engine::docker_container_stop(container_id.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // Mark build_summary end time
        // Mark build_summary.build_statue as done
        let build_summary_current_state = NewBuildSummary {
            build_target_id: build_summary_current_state_db.build_target_id,
            start_time: build_summary_current_state_db.start_time,
            end_time: Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0)),
            build_state: postgres::schema::JobState::Done,
            ..Default::default()
        };

        info!("Updating build state to done");
        let build_summary_result_done = postgres::client::build_summary_update(
            &pg_conn,
            &org_db.name,
            &repo_db.name,
            &build_target_current_state.git_hash,
            &build_target_current_state.branch,
            build_target_current_state.build_index,
            build_summary_current_state.clone(),
        )
        .expect("Unable to update build summary job state to running");

        let (_repo_db, build_target_db, build_summary_current_state_db) = (
            build_summary_result_done.0,
            build_summary_result_done.1,
            build_summary_result_done.2,
        );

        let build_metadata = BuildMetadata {
            id: repo_db.next_build_index - 1,
            build: Some(unwrapped_request),
            queue_time: Some(prost_types::Timestamp {
                seconds: build_target_db.queue_time.timestamp(),
                nanos: build_target_db.queue_time.timestamp_subsec_nanos() as i32,
            }),
            start_time: match build_summary_current_state_db.start_time {
                Some(t) => Some(prost_types::Timestamp {
                    seconds: t.timestamp(),
                    nanos: t.timestamp_subsec_nanos() as i32,
                }),
                None => None,
            },
            end_time: match build_summary_current_state_db.end_time {
                Some(t) => Some(prost_types::Timestamp {
                    seconds: t.timestamp(),
                    nanos: t.timestamp_subsec_nanos() as i32,
                }),
                None => None,
            },
            build_state: JobState::Done.into(),
            ..Default::default()
        };

        let response = Response::new(build_metadata);
        Ok(response)
    }

    async fn build_stop(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        unimplemented!();
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

        let build_stage_query = postgres::client::build_logs_get(
            &pg_conn,
            &unwrapped_request.org,
            &unwrapped_request.git_repo,
            &unwrapped_request.commit_hash,
            &unwrapped_request.branch,
            {
                match &unwrapped_request.id {
                    0 => None,
                    _ => Some(unwrapped_request.id.clone()),
                }
            },
        )
        .expect("No build stages found");

        let (mut tx, rx) = mpsc::channel(4);

        let mut build_stage_list: Vec<orbital_headers::build_meta::BuildStage> = Vec::new();
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

        tx.send(Ok(build_log_response)).await.unwrap();

        // FIXME - the protobuf struct is too complicated
        //BuildLogResponse
        // Vec<BuildRecord>

        // Check on whether the BuildTarget represents a build in progress or one that is done
        // Get the output from the database

        // TODO: This is for handling builds that are in progress
        //tokio::spawn(async move {
        //    tx.send(Ok(BuildLogResponse { ..Default::default() })).await.unwrap();

        //    println!(" /// done sending");
        //});

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
