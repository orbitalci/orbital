//use orbital_headers::orbital_types::JobTrigger;
use anyhow::anyhow;
use anyhow::Result;
use chrono::NaiveDateTime;
use config_parser::OrbitalConfig;
use git_meta::git_info;
use git_meta::{GitCommitContext, GitCredentials};
use git_url_parse::GitUrl;
use log::{debug, error, info};
use machine::{machine, transitions};
use mktemp::Temp;
use orbital_agent::build_engine;
use orbital_database::postgres;
use orbital_database::postgres::build_summary::NewBuildSummary;
use orbital_database::postgres::schema::JobTrigger;
use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
use orbital_headers::orbital_types::{JobState as PGJobState, SecretType};
use orbital_headers::secret::{secret_service_client::SecretServiceClient, SecretGetRequest};
use serde_json::Value;
use std::fs::File;
use std::io::prelude::*;
use std::path::Path;
use std::str;
use tonic::{Code, Request, Response, Status};

machine! {
    #[derive(Clone, Debug, PartialEq)]
    enum BuildState {
        Queued,
        Starting,
        Running,
        Finishing,
        Done,
        Cancelled,
        Fail,
        SystemErr
    }
}

#[derive(Clone, Debug, PartialEq)]
pub struct Step;

transitions!(BuildState,
    [
        (Queued, Step) => Starting,
        (Queued, Cancelled) => Cancelled,
        (Queued, SystemErr) => SystemErr,
        (Starting, Step) => Running,
        (Starting, Cancelled) => Cancelled,
        (Starting, SystemErr) => SystemErr,
        (Running, Step) => Running,
        (Running, Finishing) => Finishing,
        (Running, Cancelled) => Cancelled,
        (Running, SystemErr) => SystemErr,
        (Finishing, Fail) => Fail,
        (Finishing, Done) => Done
    ]
);

impl Queued {
    pub fn on_step(self, _: Step) -> Starting {
        println!("Queued -> Starting");
        Starting {}
    }

    pub fn on_cancelled(self, _: Cancelled) -> Cancelled {
        println!("Queued -> Cancelled");
        Cancelled {}
    }

    pub fn on_system_err(self, _: SystemErr) -> SystemErr {
        println!("Queued -> SystemErr");
        SystemErr {}
    }
}

impl Starting {
    pub fn on_step(self, _: Step) -> Running {
        println!("Starting -> Running");
        Running {}
    }

    pub fn on_cancelled(self, _: Cancelled) -> Cancelled {
        println!("Starting -> Cancelled");
        Cancelled {}
    }

    pub fn on_system_err(self, _: SystemErr) -> SystemErr {
        println!("Starting -> SystemErr");
        SystemErr {}
    }
}

impl Running {
    pub fn on_step(self, _: Step) -> Running {
        println!("Running -> Running");
        Running {}
    }

    pub fn on_finishing(self, _: Finishing) -> Finishing {
        println!("Running -> Finishing");
        Finishing {}
    }

    pub fn on_cancelled(self, _: Cancelled) -> Cancelled {
        println!("Running -> Cancelled");
        Cancelled {}
    }

    pub fn on_system_err(self, _: SystemErr) -> SystemErr {
        println!("Running -> SystemErr");
        SystemErr {}
    }
}

impl Finishing {
    pub fn on_fail(self, _: Fail) -> Fail {
        println!("Finishing -> Fail");
        Fail {}
    }

    pub fn on_done(self, _: Done) -> Done {
        println!("Finishing -> Done");
        Done {}
    }
}

#[derive(Clone)]
pub struct BuildContext {
    pub org: String,
    pub repo_name: String,
    pub branch: String,
    pub id: Option<i32>,
    pub hash: Option<String>,
    pub user_envs: Option<Vec<String>>,
    pub job_trigger: JobTrigger,
    pub queue_time: Option<NaiveDateTime>,
    pub start_time: Option<NaiveDateTime>,
    pub _git_clone_dir: Option<Temp>,
    pub _git_creds: Option<GitCredentials>,
    pub _git_commit_info: Option<GitCommitContext>,
    pub _build_config: Option<OrbitalConfig>,
    pub _repo_uri: Option<GitUrl>,
    _state: BuildState,
}

impl BuildContext {
    pub fn new() -> Self {
        BuildContext {
            org: "".to_string(),
            repo_name: "".to_string(),
            branch: "".to_string(),
            id: None,
            hash: None,
            user_envs: None,
            job_trigger: JobTrigger::Manual,
            queue_time: None,
            start_time: None,
            _git_clone_dir: None,
            _git_creds: None,
            _git_commit_info: None,
            _build_config: None,
            _repo_uri: None,
            _state: BuildState::queued(),
        }
    }

    pub fn add_org(mut self, org: String) -> BuildContext {
        self.org = org;
        self
    }

    pub fn add_repo_uri(mut self, repo_uri: String) -> Result<BuildContext> {
        let repo_uri_parsed =
            git_info::git_remote_url_parse(repo_uri.as_ref()).expect("Could not parse repo uri");

        self.repo_name = repo_uri_parsed.name.clone();

        self._repo_uri = Some(repo_uri_parsed);

        Ok(self)
    }

    pub fn add_repo_name(mut self, repo_name: String) -> BuildContext {
        self.repo_name = repo_name;
        self
    }

    pub fn add_branch(mut self, branch: String) -> BuildContext {
        self.branch = branch;
        self
    }

    pub fn add_id(mut self, id: i32) -> BuildContext {
        self.id = Some(id);
        self
    }

    pub fn add_hash(mut self, hash: String) -> BuildContext {
        self.hash = Some(hash);
        self
    }

    pub fn add_triggered_by(mut self, trigger: JobTrigger) -> BuildContext {
        self.job_trigger = trigger;
        self
    }

    pub fn queue(mut self) -> Result<BuildContext> {
        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        // Add build target record in db
        debug!("Adding new build target to DB");
        let build_target_result = postgres::client::build_target_add(
            &pg_conn,
            &self.org,
            &self.repo_name,
            &self.hash.clone().expect("No repo hash to target"),
            &self.branch,
            Some(self.user_envs.clone().unwrap_or_default().join("")),
            self.job_trigger.clone(),
        )
        .expect("Build target add failed");

        let (_org_db, _repo_db, build_target_db) = (
            build_target_result.0,
            build_target_result.1,
            build_target_result.2,
        );

        // Add the build id and queue timestamp BuildContext
        self.id = Some(build_target_db.build_index);
        self.queue_time = Some(build_target_db.queue_time);

        // Create a new build summary record
        debug!("Adding new build summary to DB");
        let _build_summary_result_add = postgres::client::build_summary_add(
            &pg_conn,
            &self.org,
            &self.repo_name,
            &self.hash.clone().expect("No repo hash to target"),
            &self.branch,
            self.id.clone().expect("No build id defined"),
            NewBuildSummary {
                build_target_id: build_target_db.id,
                build_state: postgres::schema::JobState::Queued,
                start_time: None,
                ..Default::default()
            },
        )
        .expect("Unable to create new build summary");

        Ok(self)
    }

    pub fn state(self) -> BuildState {
        self._state
    }

    // Change state only once then return
    // TODO: This needs to be changed to accept a channel to stream to
    pub async fn step(self) -> Result<BuildContext> {
        let mut current_step = self.clone();

        // Check for termination conditions

        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        // Check if cancelled
        match postgres::client::is_build_canceled(
            &pg_conn,
            &current_step.org,
            &current_step.repo_name,
            &current_step.hash.clone().unwrap_or_default(),
            &current_step.branch,
            current_step.id.clone().unwrap(),
        ) {
            Ok(true) => {
                info!("Build was cancelled");
                current_step._state = current_step.clone().state().on_cancelled(Cancelled {});

                // TODO: Update database

                return Ok(current_step);
            }
            Ok(false) => {
                info!("Build was not cancelled - {:?}", &self._state);
            }
            _ => {
                error!("Error checking for build cancellation");
                current_step._state = current_step.clone().state().on_system_err(SystemErr {});

                // TODO: Update database

                return Ok(current_step);
            }
        };

        //next._state = match next.clone().state() == BuildState::finishing() {
        //    true => {
        //        // If there were failures during the run, then move to failed, otherwise move to done

        //        next.clone().state().on_done(Done {})
        //    } // TODO: Manage the cleanup. Diff between pass/fail.
        //    false => {
        //        // If there are more commands to run, then step, otherwise move to finishing
        //        next.clone().state().on_step(Step)
        //    }
        //};

        let next_step = match current_step.clone().state() {
            BuildState::Queued(_) => {
                // Get secrets for cloning
                let mut next_step = current_step
                    .clone()
                    .secrets()
                    .await
                    .expect("Getting repo secrets failed");

                next_step._state.clone().on_step(Step {});
                next_step._state = BuildState::starting();
                next_step
            }
            BuildState::Starting(_) => {
                // Clone code
                // Validate orb.yml
                // Set a start time
                // Initialize stage name and step index
                let mut next_step = current_step.clone();
                next_step._state = next_step._state.clone().on_step(Step {});
                //next_step._state = BuildState::running();
                next_step
            }
            BuildState::Running(_) => {
                // Run command per step index
                // If it is a new stage, create the metadata

                // Note exit code?

                // Mark next command to run
                // If this was the last command in a stage, mark the end time

                // If the

                let mut next_step = current_step.clone();
                // Run this step once to prove loopback works
                next_step._state = next_step._state.clone().on_step(Step {});
                //next_step._state = BuildState::running();

                // DEBUG! Stepping this to Finish state immediately
                next_step._state = next_step._state.clone().on_finishing(Finishing {});
                next_step
            }
            BuildState::Finishing(_) => {
                // Set the end time for the build

                let mut next_step = current_step.clone();
                next_step._state = next_step._state.clone().on_done(Done {});
                next_step
            }
            _ => current_step.clone(),
        };

        Ok(next_step)
    }

    pub async fn secrets(mut self) -> Result<BuildContext> {
        //use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
        use crate::ServiceType;

        // Retrieve any secrets needed to clone code

        debug!("Connecting to the Code service");
        let code_client_conn_req = CodeServiceClient::connect(format!(
            "http://{}",
            super::get_service_uri(ServiceType::Code)
        ));

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

        // Build a GitCredentials struct based on the repo auth type
        // Declaring this in case we have an ssh key.
        //let mut temp_keypath = Temp::new_file().expect("Unable to create temp file");
        //let mut temp_keypath = self._ssh_key.clone().unwrap();

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
                        .clone()
                        .expect("No host defined"),
                    &self._repo_uri.clone().unwrap().name,
                );

                let secret_service_request = Request::new(SecretGetRequest {
                    org: self.org.clone().into(),
                    name: secret_name,
                    secret_type: SecretType::SshKey.into(),
                    ..Default::default()
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

                // TODO: Stop using username from Code service output

                // Replace username with the user from the code service

                let git_creds = GitCredentials::SshKey {
                    username: vault_response["username"]
                        .clone()
                        .as_str()
                        .unwrap()
                        .to_string(),
                    public_key: None,
                    private_key: vault_response["private_key"].as_str().unwrap().to_string(),
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

                git_creds
            }
            SecretType::BasicAuth => {
                info!("Basic Auth creds needed to clone");

                debug!("Connecting to the Secret service");
                let secret_client_conn_req = SecretServiceClient::connect(format!(
                    "http://{}",
                    super::get_service_uri(ServiceType::Secret)
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
                        .clone()
                        .expect("No host defined"),
                    &self._repo_uri.clone().unwrap().name,
                );

                let secret_service_request = Request::new(SecretGetRequest {
                    org: self.org.clone().into(),
                    name: secret_name,
                    secret_type: SecretType::BasicAuth.into(),
                    ..Default::default()
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

        // Don't forget to save the cloning creds
        self._git_creds = Some(git_creds);

        Ok(self)
    }

    pub fn clone_code(mut self) -> Result<BuildContext> {
        info!("Cloning code into temp directory");

        let git_repo = build_engine::clone_repo(
            format!("{}", &self._repo_uri.clone().unwrap()).as_str(),
            &self.branch,
            self._git_creds.clone().unwrap(),
        )
        .expect("Unable to clone repo");

        self._git_clone_dir = Some(git_repo.clone());

        // build stage end cloning repo.

        // Here we parse the newly cloned repo so we can get the commit message
        match git_info::get_git_info_from_path(
            git_repo.as_path(),
            &Some(self.branch.clone()),
            &Some(self.hash.clone().unwrap()),
        ) {
            Ok(git_repo_info) => {
                self._git_commit_info = Some(git_repo_info);
                Ok(self)
            }
            Err(e) => Err(e),
        }
    }

    pub fn add_build_config_from_path(mut self) -> Result<BuildContext> {
        match (self._git_clone_dir.clone(), self._build_config.clone()) {
            (Some(code), Some(_config)) => {
                // Re-parse and re-set the config
                info!("Re-loading build config from cloned code");
                let c = build_engine::load_orb_config(Path::new(&format!(
                    "{}/{}",
                    &code.as_path().display(),
                    "orb.yml",
                )))
                .expect("Unable to load orb.yml");

                self._build_config = Some(c);

                Ok(self)
            }
            (Some(code), None) => {
                // Parse and re-set the config
                info!("Build config not yet set. Loading build config from cloned code");
                let c = build_engine::load_orb_config(Path::new(&format!(
                    "{}/{}",
                    &code.as_path().display(),
                    "orb.yml",
                )))
                .expect("Unable to load orb.yml");

                self._build_config = Some(c);

                Ok(self)
            }
            (None, _) => Err(anyhow!("Code is not cloned")),
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
}
