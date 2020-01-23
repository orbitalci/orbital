use orbital_headers::build_meta::{
    build_service_server::BuildService,
    BuildLogResponse,
    BuildMetadata, //BuildRecord, BuildStage,
    BuildSummaryRequest,
    BuildSummaryResponse,
    BuildTarget,
};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
use orbital_headers::orbital_types::{JobState, SecretType};
use orbital_headers::secret::{secret_service_client::SecretServiceClient, SecretGetRequest};

use tonic::{Request, Response, Status};

//use tokio::sync::mpsc;

use crate::OrbitalServiceError;
use agent_runtime::build_engine;
use git_meta::GitCredentials;

use super::{OrbitalApi, ServiceType};

use log::debug;

use prost_types::Timestamp;
use std::path::Path;
use std::time::{Duration, SystemTime};

use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;

use futures::Stream;
use std::pin::Pin;

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

        let start_timestamp = Timestamp {
            seconds: SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs() as i64,
            nanos: 0,
        };

        // Git clone for provider ( uri, branch, commit )
        let unwrapped_request = request.into_inner();
        debug!("Received request: {:?}", &unwrapped_request);

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
        // e.g.: default_org/github.com/level11consulting/orbitalci
        let request_payload = Request::new(GitRepoGetRequest {
            org: unwrapped_request.org.clone().into(),
            git_provider: unwrapped_request.clone().git_provider,
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

        let git_creds = match &code_service_response.secret_type.into() {
            SecretType::Unspecified => {
                debug!("No secret needed to clone. Public repo");

                GitCredentials::Public
            }
            SecretType::SshKey => {
                debug!("SSH key needed to clone");

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

                // This is kind of a hack until I clean up git repo uri duplication starting from the client-side
                // I want the vault path for git repo keys to be treated like any other secret for an org
                // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                // Where the secret name is the git repo url
                // e.g., "github.com/level11consulting/orbitalci"

                let secret_name = format!(
                    "{}/{}",
                    &unwrapped_request.git_provider, &unwrapped_request.git_repo,
                );

                //let vault_secret_path = agent_runtime::vault::orb_vault_path(
                //    &unwrapped_request.org.clone(),
                //    &secret_name,
                //    &SecretType::SshKey.to_string(),
                //);

                //let vault_secret_path = format!(
                //    "orbital/{}/{}/{}/{}",
                //    &unwrapped_request.org,
                //    SecretType::SshKey.to_string().to_lowercase(),
                //    &unwrapped_request.git_provider,
                //    &unwrapped_request.git_repo,
                //);

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

                // Write ssh key to temp file
                debug!("Writing incoming ssh key to temp file");
                let mut file = File::create(temp_keypath.as_path())?;
                let mut _contents = String::new();
                let _ = file.write_all(&secret_service_response.data);

                // Replace username with the user from the code service
                let git_creds = GitCredentials::SshKey {
                    username: code_service_response.user,
                    public_key: None,
                    private_key: temp_keypath.as_path(),
                    passphrase: None,
                };

                git_creds
            }
            SecretType::BasicAuth => {
                debug!("Userpass needed to clone");
                let git_creds = GitCredentials::UserPassPlaintext {
                    username: "git".to_string(),
                    password: "fakepassword".to_string(),
                };

                git_creds
            }
            _ => panic!(
                "We only support public repos, or private repo auth with sshkeys or basic auth"
            ),
        };

        //
        debug!("Cloning code into temp directory");
        let git_repo = build_engine::clone_repo(
            &unwrapped_request.remote_uri,
            &unwrapped_request.branch,
            git_creds,
        )
        .expect("Unable to clone repo");

        debug!("Loading orb.yml from path {:?}", &git_repo.as_path());
        let config = build_engine::load_orb_config(Path::new(&format!(
            "{}/{}",
            &git_repo.as_path().display(),
            "orb.yml"
        )))
        .expect("Unable to load orb.yml");

        debug!("Pulling container: {:?}", config.image.clone());
        match build_engine::docker_container_pull(config.image.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // TODO: Inject the dynamic build env vars here
        let envs_vec = agent_runtime::parse_envs_input(&None);
        let vols_vec = agent_runtime::parse_volumes_input(&None);

        // Create a new container
        debug!("Creating container");
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
        debug!("Start container");
        match build_engine::docker_container_start(container_id.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        // TODO: Make sure tests try to exec w/o starting the container
        // Exec into the new container
        debug!("Sending commands into container");
        match build_engine::docker_container_exec(container_id.as_str(), config.command) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        debug!("Stopping the container");
        match build_engine::docker_container_stop(container_id.as_str()) {
            Ok(ok) => ok, // The successful result doesn't matter
            Err(e) => return Err(OrbitalServiceError::new(&e.to_string()).into()),
        };

        let end_timestamp = Timestamp {
            seconds: SystemTime::now()
                .duration_since(SystemTime::UNIX_EPOCH)
                .unwrap()
                .as_secs() as i64,
            nanos: 0,
        };

        let build_metadata = BuildMetadata {
            build: Some(unwrapped_request),
            start_time: Some(start_timestamp),
            end_time: Some(end_timestamp),
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

    type BuildLogsStream =
        Pin<Box<dyn Stream<Item = Result<BuildLogResponse, Status>> + Send + Sync + 'static>>;

    async fn build_logs(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<tonic::Response<Self::BuildLogsStream>, tonic::Status> {
        unimplemented!()
        //// Get the container info so we can call the docker api

        //let (mut tx, rx) = mpsc::channel(4);

        //tokio::spawn(async move {
        //    let build_log_response = BuildLogResponse {
        //        id: 0,
        //        records: Vec::new(),
        //    };

        //    tx.send(Ok(build_log_response)).await.unwrap()

        //    // Determine if there are logs in the database and fetch those first
        //    // Probably send those logs, unless we specifically don't want to

        //    // If the build is still running, then we want to go to the live logs

        //    // The agent runtime log wrapper needs to provide the build stage info
        //});

        //Ok(Response::new(rx))
    }

    async fn build_remove(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        unimplemented!();
    }

    async fn build_summary(
        &self,
        _request: Request<BuildSummaryRequest>,
    ) -> Result<Response<BuildSummaryResponse>, Status> {
        unimplemented!();
    }
}
