use crate::orbital_database::postgres;
use crate::orbital_database::postgres::build_stage::NewBuildStage;
use crate::orbital_utils::config_parser::OrbitalConfig;
use crate::orbital_utils::orbital_agent::{self, build_engine};
use chrono::{NaiveDateTime, Utc};
use color_eyre::eyre::Result;
use git_meta::{GitCommitMeta, GitCredentials, GitRepo};
use git_url_parse::GitUrl;
use tracing::{debug, error, info};
use std::fs::File;
use std::io::prelude::*;

use crate::orbital_database::postgres::build_summary::NewBuildSummary;
use crate::orbital_database::postgres::schema::JobTrigger;
use crate::orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
use crate::orbital_headers::orbital_types::SecretType;
use crate::orbital_headers::secret::{
    secret_service_client::SecretServiceClient, SecretGetRequest,
};
use crate::orbital_services::get_service_uri;
use crate::orbital_utils::exec_runtime::docker::OrbitalContainerSpec;
use serde_json::Value;
use std::path::{Path, PathBuf};
use std::str;
use std::time::Duration;
use tokio::sync::mpsc;
use tonic::Request;

use super::database_helper::DbHelper;
use crate::orbital_services::state_machine::build_state::BuildState;
use crate::orbital_services::state_machine::build_state::{
    Cancelled, Done, Finishing, Step, SystemErr,
};
use crate::orbital_services::DEFAULT_BUILD_TIMEOUT;

#[derive(Clone)]
pub struct BuildContext {
    pub org: String,
    pub repo_name: String,
    pub branch: String,
    pub working_dir: PathBuf,
    pub private_key: PathBuf,
    pub id: Option<i32>,
    pub hash: Option<String>,
    pub user_envs: Option<Vec<String>>,
    pub job_trigger: JobTrigger,
    pub queue_time: Option<NaiveDateTime>,
    pub build_start_time: Option<NaiveDateTime>,
    pub build_end_time: Option<NaiveDateTime>,
    pub stage_start_time: Option<NaiveDateTime>,
    pub stage_end_time: Option<NaiveDateTime>,
    pub build_stage_name: String,
    pub build_cur_stage_logs: String,
    pub _git_creds: Option<GitCredentials>,
    pub _git_commit_info: Option<GitCommitMeta>,
    pub _build_config: Option<OrbitalConfig>,
    pub _container_id: Option<String>,
    pub _repo_uri: Option<GitUrl>,
    _build_hostname: String,
    _db_build_target_id: i32,
    pub _db_build_summary_id: i32,
    pub _db_build_cur_stage_id: i32,
    _build_progress_marker: (usize, usize),
    _state: BuildState,
}

impl BuildContext {
    pub fn new() -> Self {
        BuildContext {
            org: String::new(),
            repo_name: String::new(),
            branch: String::new(),
            working_dir: PathBuf::new(),
            private_key: PathBuf::new(),
            id: None,
            hash: None,
            user_envs: None,
            job_trigger: JobTrigger::Manual,
            queue_time: None,
            build_start_time: None,
            build_end_time: None,
            stage_start_time: None,
            stage_end_time: None,
            build_stage_name: String::from("Queued"),
            build_cur_stage_logs: String::new(),
            _git_creds: None,
            _git_commit_info: None,
            _build_config: None,
            _container_id: None,
            _db_build_target_id: 0,
            _db_build_summary_id: 0,
            _db_build_cur_stage_id: 0,
            _build_progress_marker: (0, 0),
            _repo_uri: None,
            _build_hostname: hostname::get().unwrap().into_string().unwrap(),

            _state: BuildState::queued(),
        }
    }

    pub fn add_org(mut self, org: String) -> BuildContext {
        self.org = org;
        self
    }

    pub fn add_repo_uri(mut self, repo_uri: String) -> Result<BuildContext> {
        let repo_uri_parsed = GitUrl::parse(repo_uri.as_ref()).expect("Could not parse repo uri");

        self.repo_name = repo_uri_parsed.name.clone();

        self._repo_uri = Some(repo_uri_parsed);

        Ok(self)
    }

    //pub fn add_repo_name(mut self, repo_name: String) -> BuildContext {
    //    self.repo_name = repo_name;
    //    self
    //}

    pub fn add_branch(mut self, branch: String) -> BuildContext {
        self.branch = branch;
        self
    }

    //pub fn add_id(mut self, id: i32) -> BuildContext {
    //    self.id = Some(id);
    //    self
    //}

    pub fn add_hash(mut self, hash: String) -> BuildContext {
        self.hash = Some(hash);
        self
    }

    pub fn add_triggered_by(mut self, trigger: JobTrigger) -> BuildContext {
        self.job_trigger = trigger;
        self
    }

    pub fn add_working_dir(mut self, working_dir: PathBuf) -> BuildContext {
        self.working_dir = working_dir;
        self
    }

    pub fn add_private_key(mut self, private_key: PathBuf) -> BuildContext {
        self.private_key = private_key;
        self
    }

    pub fn queue(mut self) -> Result<BuildContext> {
        // Add build target record in db
        debug!("Adding new build target to DB");
        let build_target_result =
            DbHelper::build_target_add(&self).expect("Build target add failed");

        let (_org_db, _repo_db, build_target_db) = (
            build_target_result.0,
            build_target_result.1,
            build_target_result.2,
        );

        // Add the build id and queue timestamp BuildContext
        self.id = Some(build_target_db.build_index);
        self._db_build_target_id = build_target_db.id;
        self.queue_time = Some(build_target_db.queue_time);

        // Create a new build summary record
        debug!("Adding new build summary to DB");
        let build_summary_result_add = DbHelper::build_summary_add(
            &self,
            NewBuildSummary {
                build_target_id: build_target_db.id,
                build_state: postgres::schema::JobState::Queued,
                start_time: None,
                ..Default::default()
            },
        )
        .expect("Unable to create new build summary");

        // Save build summary id
        let (_repo_db, _build_target_db, build_summary_db) = (
            build_summary_result_add.0,
            build_summary_result_add.1,
            build_summary_result_add.2,
        );

        self._db_build_summary_id = build_summary_db.id;

        Ok(self)
    }

    pub fn state(self) -> BuildState {
        self._state
    }

    //pub fn get_context(self) -> BuildContext {
    //    self
    //}

    pub async fn step(self, caller_tx: &mpsc::UnboundedSender<String>) -> Result<BuildContext> {
        // Check for termination conditions

        // Check if cancelled
        let _cancel_check = match DbHelper::is_build_cancelled(&self) {
            Ok(true) => {
                info!("Build was cancelled");
                let mut next_step = self.clone();
                next_step._state = self.clone().state().on_cancelled(Cancelled {});

                // TODO: Update database

                return Ok(next_step);
            }
            Ok(false) => {
                info!("Build was not cancelled - {:?}", &self._state);
            }
            _ => {
                error!("Error checking for build cancellation");
                let mut next_step = self.clone();
                next_step._state = self.clone().state().on_system_err(SystemErr {});

                // TODO: Update database

                return Ok(next_step);
            }
        };

        let next_step = match self.clone().state() {
            BuildState::Queued(_) => {
                // Get secrets for cloning
                let mut next_step = self
                    .clone()
                    .secrets()
                    .await
                    .expect("Getting repo secrets failed");

                let _ = caller_tx.send("Stream: Queued -> Starting".to_string());

                next_step._state = next_step._state.clone().on_step(Step {});
                next_step.build_stage_name = "Starting".to_string();
                next_step
            }
            BuildState::Starting(_) => {
                let mut next_step = self.clone();

                // Set build start time
                next_step.build_start_time =
                    Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0));

                // Update DB
                info!("Updating build state to starting in DB");
                let _build_summary_result_start = DbHelper::build_summary_update(
                    &self,
                    NewBuildSummary {
                        build_target_id: next_step._db_build_target_id,
                        build_state: postgres::schema::JobState::Starting,
                        start_time: next_step.build_start_time,
                        ..Default::default()
                    },
                )
                .expect("Unable to update build summary job state to starting");

                // Clone code
                next_step = next_step.clone_code().expect("Cloning code failed");

                // Validate orb.yml if not yet done
                if let Some(_config) = next_step._build_config.clone() {
                    info!("Config file was passed in. Not reloading");
                } else {
                    next_step = next_step
                        .add_build_config_from_path()
                        .expect("Loading config from cloned code failed");
                }

                let mut global_env_vars = next_step.clone()._internal_env_vars();

                if let Some(config_env_globals) = next_step._build_config.clone().unwrap().env {
                    global_env_vars.extend(config_env_globals);
                }

                // TODO: Use this spec when we can pre-populate the entire build info from config
                let build_container_spec = OrbitalContainerSpec {
                    name: Some(orbital_agent::generate_unique_build_id(
                        &next_step.org,
                        &next_step.repo_name,
                        &next_step.hash.clone().unwrap(),
                        &format!("{}", next_step.id.unwrap()),
                    )),
                    image: next_step._build_config.clone().unwrap().image,
                    command: Vec::new(), // TODO: Populate this field

                    env_vars: Some(global_env_vars.iter().map(|s| s.as_ref()).collect()),
                    volumes: crate::orbital_utils::exec_runtime::parse_volumes_input(&None),
                    timeout: Some(Duration::from_secs(DEFAULT_BUILD_TIMEOUT)),
                };

                // Pull container
                info!(
                    "Pulling container: {:?}",
                    build_container_spec.image.clone()
                );

                // I guess here's where I read from the channel?
                let mut stream = build_engine::docker_container_pull(build_container_spec.clone())
                    .await
                    .unwrap();

                while let Some(response) = stream.recv().await {
                    info!("PULL OUTPUT: {:?}", response["status"].clone().as_str());
                    let output = format!("{}\n", response["status"].clone().as_str().unwrap());

                    let _ = caller_tx.send(output.clone());
                }

                // Start the container

                // Create a new container
                info!("Creating container");
                next_step._container_id =
                    build_engine::docker_container_create(&build_container_spec)
                        .await
                        .ok();

                // Start a docker container
                info!(
                    "Start container {:?}",
                    &next_step._container_id.clone().expect(
                        "Could not create container - Does the container name already exist?"
                    )
                );
                let _ =
                    build_engine::docker_container_start(&next_step._container_id.clone().unwrap())
                        .await
                        .unwrap();

                let _ = caller_tx.send("Stream: Starting -> Running".to_string());

                next_step._state = next_step._state.clone().on_step(Step {});
                next_step.build_stage_name = "Running".to_string();
                next_step
            }
            BuildState::Running(_) => {
                // Mark timestamp
                let step_start_timestamp = NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0);

                // Note exit code?

                let mut next_step = self.clone();

                // Mark the start of the build stage if it hasn't been set yet
                if next_step.stage_start_time.is_none() {
                    next_step.stage_start_time = Some(step_start_timestamp);

                    let build_stage_start = DbHelper::build_stage_add(
                        &self,
                        NewBuildStage {
                            build_summary_id: next_step._db_build_summary_id,
                            stage_name: Some(next_step.build_stage_name),
                            build_host: Some(next_step._build_hostname.clone()),
                            ..Default::default()
                        },
                    )
                    .expect("Unable to add new build stage in db");

                    let (_build_target_db, _build_summary_db, build_stage_db) = (
                        build_stage_start.0,
                        build_stage_start.1,
                        build_stage_start.2,
                    );

                    next_step._db_build_cur_stage_id = build_stage_db.id;

                    let _build_summary_result_running = DbHelper::build_summary_update(
                        &self,
                        NewBuildSummary {
                            build_target_id: next_step._db_build_target_id,
                            build_state: postgres::schema::JobState::Running,
                            ..Default::default()
                        },
                    )
                    .expect("Unable to update build summary job state to running");
                }

                let c = next_step._build_config.clone().unwrap();

                let stage_index = next_step._build_progress_marker.0;
                let command_index = next_step._build_progress_marker.1;

                println!(
                    "Stage index:{} Command index:{}",
                    stage_index, command_index
                );

                println!("{:?}", c.stages[stage_index].command[command_index]);
                next_step.build_stage_name = c.stages[stage_index]
                    .name
                    .clone()
                    .unwrap_or(format!("Stage {}", stage_index));

                let mut stream = build_engine::docker_container_exec(
                    next_step._container_id.clone().unwrap(),
                    vec![c.stages[stage_index].command[command_index].clone()],
                )
                .await
                .unwrap();

                // Stream build output to client, and save to BuildContext
                while let Some(response) = stream.recv().await {
                    let _ = caller_tx.send(response.clone());
                    next_step
                        .build_cur_stage_logs
                        .push_str(response.clone().as_str());
                }

                let step_end_timestamp = NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0);

                // Update build progress marker
                if c.stages
                    .get(stage_index)
                    .unwrap()
                    .command
                    .get(command_index + 1)
                    .is_some()
                {
                    // Set the next stage name
                    next_step.build_stage_name = c
                        .stages
                        .get(stage_index)
                        .unwrap()
                        .name
                        .clone()
                        .unwrap_or(format!("stage #{}", stage_index));

                    // Next command in the stage
                    next_step._build_progress_marker = (stage_index, command_index + 1);

                    let _ = caller_tx.send("Stream: Running -> Running (Next command)".to_string());
                    next_step._state = next_step._state.clone().on_step(Step {});

                // Note to future self: We don't want to add stage logs until we're done with the stage
                } else if c.stages.get(stage_index + 1).is_some() {
                    // First command of the next stage
                    next_step._build_progress_marker = (stage_index + 1, 0);

                    let _ = caller_tx.send("Stream: Running -> Running (Next stage)".to_string());
                    next_step._state = next_step._state.clone().on_step(Step {});
                    next_step.stage_end_time = Some(step_end_timestamp);

                    // End stage
                    info!("Marking end of build stage");
                    let _build_stage_end = DbHelper::build_stage_update(
                        &self,
                        NewBuildStage {
                            build_summary_id: next_step._db_build_summary_id,
                            stage_name: Some(next_step.build_stage_name.clone()),
                            start_time: next_step.stage_start_time.unwrap(),
                            end_time: next_step.stage_end_time,
                            build_host: Some(next_step._build_hostname.clone()),
                            output: Some(next_step.build_cur_stage_logs.clone()),
                            ..Default::default()
                        },
                    );

                    // Reset stage timestamps
                    next_step.stage_start_time = None;
                    next_step.stage_end_time = None;

                    // Reset stage logs
                    next_step.build_cur_stage_logs = String::new();
                } else {
                    // This was the last command

                    let _ = caller_tx.send("Stream: Running -> Finishing".to_string());
                    next_step._state = next_step._state.clone().on_finishing(Finishing {});
                    next_step.stage_end_time = Some(step_end_timestamp);

                    // Update DB to Finishing

                    // End stage
                    info!("Marking end of build stage");
                    let _build_stage_end = DbHelper::build_stage_update(
                        &self,
                        NewBuildStage {
                            build_summary_id: next_step._db_build_summary_id,
                            stage_name: Some(next_step.build_stage_name.clone()),
                            start_time: next_step.stage_start_time.unwrap(),
                            end_time: next_step.stage_end_time,
                            build_host: Some(next_step._build_hostname.clone()),
                            output: Some(next_step.build_cur_stage_logs.clone()),
                            ..Default::default()
                        },
                    );

                    let _build_summary_result_running = DbHelper::build_summary_update(
                        &self,
                        NewBuildSummary {
                            build_target_id: next_step._db_build_target_id,
                            start_time: next_step.build_start_time,
                            end_time: next_step.build_end_time,
                            build_state: postgres::schema::JobState::Finishing,
                        },
                    )
                    .expect("Unable to update build summary job state to finishing");
                } // End of build

                // End of build handling

                next_step
            }
            BuildState::Finishing(_) => {
                let mut next_step = self.clone();

                // Set build end time
                next_step.build_end_time =
                    Some(NaiveDateTime::from_timestamp(Utc::now().timestamp(), 0));

                info!("Stopping the container");
                let _ = build_engine::docker_container_stop(
                    next_step._container_id.clone().unwrap().as_str(),
                )
                .await
                .unwrap();

                next_step.build_stage_name = "Finishing".to_string();
                let _ = caller_tx.send("Stream: Finishing -> Done".to_string());

                next_step._state = next_step._state.clone().on_done(Done {});
                next_step.build_stage_name = "Done".to_string();

                let _build_summary_result_running = DbHelper::build_summary_update(
                    &self,
                    NewBuildSummary {
                        build_target_id: next_step._db_build_target_id,
                        start_time: next_step.build_start_time,
                        end_time: next_step.build_end_time,
                        build_state: postgres::schema::JobState::Done,
                    },
                )
                .expect("Unable to update build summary job state to done");

                next_step
            }
            _ => self.clone(),
        };

        Ok(next_step)
    }

    pub async fn secrets(mut self) -> Result<BuildContext> {
        use crate::orbital_services::ServiceType;

        // Retrieve any secrets needed to clone code

        debug!("Connecting to the Code service");
        let code_client_conn_req =
            CodeServiceClient::connect(format!("http://{}", get_service_uri(ServiceType::Code)));

        let mut code_client = code_client_conn_req.await.unwrap();

        debug!("Building request to Code service for git repo info");

        // Request: org/git_provider/name
        // e.g.: org_name/github.com/orbitalci/orbital
        let request_payload = Request::new(GitRepoGetRequest {
            org: self.org.clone(),
            name: self.repo_name.clone(),
            uri: format!("{}", self._repo_uri.clone().unwrap()),
            ..Default::default()
        });

        debug!("Payload: {:?}", &request_payload);

        debug!("Sending request to Code service for git repo");

        let code_service_request = code_client.git_repo_get(request_payload);
        let code_service_response = code_service_request.await.unwrap().into_inner();

        let git_creds = match &code_service_response.secret_type.into() {
            SecretType::Unspecified => {
                // TODO: Call secret service and get a username
                info!("No secret needed to clone. Public repo");

                None
            }
            SecretType::SshKey => {
                info!("SSH key needed to clone");

                debug!("Connecting to the Secret service");
                let secret_client_conn_req = SecretServiceClient::connect(format!(
                    "http://{}",
                    get_service_uri(ServiceType::Secret)
                ));

                let mut secret_client = secret_client_conn_req.await.unwrap();

                debug!("Building request to Secret service for git repo ");

                // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                // Where the secret name is the git repo url
                // e.g., "github.com/level11consulting/orbitalci"

                let secret_name = format!(
                    "{}/{}",
                    &self
                        ._repo_uri
                        .clone()
                        .unwrap()
                        .host
                        .expect("No host defined"),
                    &self._repo_uri.clone().unwrap().name,
                );

                let secret_service_request = Request::new(SecretGetRequest {
                    org: self.org.clone(),
                    name: secret_name,
                    secret_type: SecretType::SshKey.into(),
                });

                debug!("Secret request: {:?}", &secret_service_request);

                let secret_service_response = secret_client
                    .secret_get(secret_service_request)
                    .await
                    .unwrap()
                    .into_inner();

                debug!("Secret get response: {:?}", &secret_service_response);

                // TODO: Deserialize vault data into hashmap.
                let vault_response: Value =
                    serde_json::from_str(str::from_utf8(&secret_service_response.data).unwrap())
                        .expect("Unable to read json data from Vault");

                // Write ssh key to temp file
                info!("Writing incoming ssh key to GitCredentials::SshKey");
                let mut file = match File::create(&self.private_key) {
                    Err(why) => panic!("couldn't create {}: {}", &self.private_key.display(), why),
                    Ok(file) => file,
                };

                match file.write_all(vault_response["private_key"].as_str().unwrap().as_bytes()) {
                    Err(why) => {
                        panic!("couldn't write to {}: {}", &self.private_key.display(), why)
                    }
                    Ok(_) => println!("successfully wrote to {}", &self.private_key.display()),
                }

                let git_creds = GitCredentials::SshKey {
                    username: vault_response["username"]
                        .clone()
                        .as_str()
                        .unwrap()
                        .to_string(),
                    public_key: None,
                    private_key: self.private_key.clone(),
                    passphrase: None,
                };

                // Add username to git_parsed_uri for cloning
                let mut updated_git_url = self._repo_uri.unwrap();
                updated_git_url.user = Some(
                    vault_response["username"]
                        .clone()
                        .as_str()
                        .unwrap()
                        .to_string(),
                );
                self._repo_uri = Some(updated_git_url);

                debug!("Git Creds: {:?}", &git_creds);

                Some(git_creds)
            }
            SecretType::BasicAuth => {
                info!("Basic Auth creds needed to clone");

                debug!("Connecting to the Secret service");
                let secret_client_conn_req = SecretServiceClient::connect(format!(
                    "http://{}",
                    get_service_uri(ServiceType::Secret)
                ));
                let mut secret_client = secret_client_conn_req.await.unwrap();

                debug!("Building request to Secret service for git repo ");

                // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                // Where the secret name is the git repo url
                // e.g., "github.com/orbitalci/orbital"

                let secret_name = format!(
                    "{}/{}",
                    &self
                        ._repo_uri
                        .clone()
                        .unwrap()
                        .host
                        .expect("No host defined"),
                    &self._repo_uri.clone().unwrap().name,
                );

                let secret_service_request = Request::new(SecretGetRequest {
                    org: self.org.clone(),
                    name: secret_name,
                    secret_type: SecretType::BasicAuth.into(),
                });

                debug!("Secret request: {:?}", &secret_service_request);

                let secret_service_response = secret_client
                    .secret_get(secret_service_request)
                    .await
                    .unwrap()
                    .into_inner();

                debug!("Secret get response: {:?}", &secret_service_response);

                // TODO: Deserialize vault data into hashmap.
                let vault_response: Value =
                    serde_json::from_str(str::from_utf8(&secret_service_response.data).unwrap())
                        .expect("Unable to read json data from Vault");

                // Replace username with the user from the code service
                let git_creds = GitCredentials::UserPassPlaintext {
                    username: vault_response["username"].as_str().unwrap().to_string(),
                    password: vault_response["password"].as_str().unwrap().to_string(),
                };

                debug!("Git Creds: {:?}", &git_creds);
                Some(git_creds)
            }
            _ => panic!(
                "We only support public repos, or private repo auth with sshkeys or basic auth"
            ),
        };

        // Don't forget to save the cloning creds
        self._git_creds = git_creds;

        Ok(self)
    }

    pub fn clone_code(mut self) -> Result<BuildContext> {
        info!(
            "Cloning code into temp directory - {}",
            format!("{}", &self._repo_uri.clone().unwrap()).as_str()
        );

        let _clone_res = build_engine::clone_repo(
            format!("{}", &self._repo_uri.clone().unwrap()).as_str(),
            Some(&self.branch),
            self._git_creds.clone(),
            self.working_dir.as_ref(),
        )
        .expect("Unable to clone repo");

        //TODO: Add debug here to list files from self.working_dir

        // Here we parse the newly cloned repo so we can get the commit message

        let git_repo = GitRepo::open(
            self.working_dir.clone(),
            Some(self.branch.clone()),
            self.hash.clone(),
        )
        .expect("Unable to open cloned repo");

        self._git_commit_info = git_repo.head;

        //match git_info::get_git_info_from_path(
        //    self.working_dir.clone().as_path(),
        //    &Some(self.branch.clone()),
        //    &Some(self.hash.clone().unwrap()),
        //) {
        //    Ok(git_repo_info) => {
        //        self._git_commit_info = Some(git_repo_info);
        //    }
        //    Err(e) => {
        //        error!("Failed to parse metadata from repo");
        //        return Err(e);
        //    }
        //};

        Ok(self)
    }

    pub fn add_build_config_from_path(mut self) -> Result<BuildContext> {
        match self._build_config.clone() {
            Some(_config) => {
                // Re-parse and re-set the config
                info!("Re-loading build config from cloned code");
                let c = build_engine::load_orb_config(Path::new(&format!(
                    "{}/{}",
                    self.working_dir.as_path().display(),
                    "orb.yml",
                )))
                .expect("Unable to load orb.yml");

                self._build_config = Some(c);

                Ok(self)
            }
            None => {
                // Parse and re-set the config
                info!("Build config not yet set. Loading build config from cloned code");
                let c = build_engine::load_orb_config(Path::new(&format!(
                    "{}/{}",
                    self.working_dir.as_path().display(),
                    "orb.yml",
                )))
                .expect("Unable to load orb.yml");

                self._build_config = Some(c);

                Ok(self)
            }
        }
    }

    pub fn add_build_config_from_string(mut self, config: String) -> Result<BuildContext> {
        match build_engine::load_orb_config_from_str(&config) {
            Ok(c) => {
                self._build_config = Some(c);
                Ok(self)
            }
            Err(e) => Err(e),
        }
    }

    fn _internal_env_vars(self) -> Vec<String> {
        // Defining internal env vars here
        let orb_org_env = format!("ORBITAL_ORG={}", self.org);

        let orb_repo_env = format!("ORBITAL_REPOSITORY={}", self.repo_name);

        let orb_build_number_env = format!(
            "ORBITAL_BUILD_NUMBER={}",
            self.id.expect("Build number not yet set")
        );

        let orb_commit_env = format!(
            "ORBITAL_COMMIT={}",
            self.hash.clone().expect("Git hash info unavailable")
        );

        let orb_commit_short_env = format!(
            "ORBITAL_COMMIT_SHORT={}",
            &self.hash.clone().expect("Git hash info unavailable")[0..6]
        );

        let orb_commit_message = format!(
            "ORBITAL_COMMIT_MSG={}",
            self._git_commit_info
                .expect("Git commit info unavailable")
                .message
                .unwrap_or_default()
        );

        let orbital_env_vars_vec = vec![
            orb_org_env,
            orb_repo_env,
            orb_build_number_env,
            orb_commit_env,
            orb_commit_short_env,
            orb_commit_message,
        ];

        orbital_env_vars_vec
    }
}
