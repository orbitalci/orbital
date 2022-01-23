use tokio::time::{self, Duration};
use tracing::{debug, info};

use crate::orbital_headers::code::{
    code_service_client::CodeServiceClient, GitRepoListRequest, GitRepoRemoteBranchHead,
    GitRepoRemoteBranchHeadList, GitRepoUpdateRequest,
};
use crate::orbital_headers::orbital_types::SecretType;
use crate::orbital_headers::organization::organization_service_client::OrganizationServiceClient;
use crate::orbital_headers::secret::{
    secret_service_client::SecretServiceClient, SecretGetRequest,
};
use crate::orbital_services::ORB_DEFAULT_URI;
use git_event;
use git_url_parse::GitUrl;
use hex;
use serde_json::Value;
use std::str;
use tonic::Request;

// For writing ssh key to file
use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;

// For starting new build
use crate::orbital_headers::build_meta::{build_service_client::BuildServiceClient, BuildTarget};

use crate::orbital_headers::orbital_types::JobTrigger;

use git_meta::GitCredentials;

pub async fn poll_for_new_commits(poll_freq: u8) {
    tokio::spawn(async move {
        info!("Waiting a moment for server to start before polling");
        time::sleep(Duration::from_secs(5)).await;

        loop {
            debug!("Start of poll loop");

            // Gather all orgs

            info!("Collecting list of orgs");
            let mut client =
                OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI))
                    .await
                    .unwrap();
            let request = Request::new(());

            // TODO: We should filter this into only orgs that are enabled
            let orgs: Vec<String> = client
                .org_list(request)
                .await
                .unwrap()
                .into_inner()
                .orgs
                .into_iter()
                .map(|o| o.name)
                .collect();

            debug!("Orgs: {:?}", orgs);

            // Gather all repos per org

            let mut repos_response;

            info!("Looping through orgs to collect their repos");
            for org in orgs {
                let mut code_client =
                    CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI))
                        .await
                        .unwrap();
                let request = Request::new(GitRepoListRequest {
                    org: org.clone(),
                    ..Default::default()
                });

                info!("Collecting repos for org: {}", &org);
                repos_response = code_client
                    .git_repo_list(request)
                    .await
                    .unwrap()
                    .into_inner()
                    .git_repos;

                debug!("repos: {:?}", repos_response);

                // Disconnect from CodeService
                //drop(client);

                for r in repos_response {
                    // NEED: Last known commit from DB
                    // FIXME: Yikes.
                    let last_known_heads = r
                        .remote_branch_heads
                        .expect("No last known commits in DB")
                        .remote_branch_heads;

                    // FIXME: If auth required, gather auth from SecretService
                    // Load up a list of all the repos, and the path to any secrets for cloning
                    // Per branch, compare the last known remote HEAD commit, and the current remote commit.

                    // Get the uri
                    let mut uri = GitUrl::parse(&r.uri).expect("Git URI parsing failed");

                    let temp_ssh_privkey =
                        Temp::new_file().expect("Unable to create a new file for ssh key");

                    let git_creds: Option<GitCredentials> = match r.secret_type.into() {
                        SecretType::Unspecified => {
                            None
                        },
                        SecretType::SshKey => {
                            // Connect to SecretService so we can pull auth info for cloning if necessary
                            let mut secret_client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI))
                                .await
                                .unwrap();

                            // vault path pattern: /secret/orbital/<org name>/<secret type>/<secret name>
                            // Where the secret name is the git repo url
                            // e.g., "github.com/orbitalci/orbital"

                            let secret_name = format!(
                                "{}/{}",
                                &uri
                                    .clone()
                                    .host
                                    .clone()
                                    .expect("No host defined"),
                                &uri.clone().name,
                            );

                            let secret_service_request = Request::new(SecretGetRequest {
                                org: org.clone(),
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


                            // Grab the relevant info from the vault response
                            let user = vault_response["username"].clone().as_str().unwrap().to_string();
                            let private_key = vault_response["private_key"].as_str().unwrap().to_string();

                            // Write the private key to temp file
                            let mut privkey_file = File::create(temp_ssh_privkey.as_path()).expect("Unable to open file for writing");
                            privkey_file.write_all(private_key.as_bytes()).expect("Error writing key to file");


                            // Add the username to the uri. Needed for cloning
                            uri.user = Some(user.clone());


                            Some(GitCredentials::SshKey{
                                public_key: None,
                                private_key: temp_ssh_privkey.to_path_buf(),
                                username: user,
                                passphrase: None,
                            })
                        },
                        SecretType::BasicAuth => {

                            // TODO: Translate the logic for BasicAuth creds out of vault from build_context.rs
                            unimplemented!("Basic auth not yet implemented")
                        },
                        _ => panic!(
                            "We only support public repos, or private repo auth with sshkeys or basic auth"
                        ),
                    };

                    //let repo_watcher = git_event::GitRepoWatchHandler::new(&r.uri)
                    //    .unwrap()
                    //    .with_shallow_clone(true);

                    let mut repo_watcher = if let Some(creds) = git_creds {
                        git_event::GitRepoWatchHandler::new(format!("{}", uri))
                            .unwrap()
                            .with_credentials(Some(creds))
                            .with_shallow_clone(true)
                    } else {
                        git_event::GitRepoWatchHandler::new(format!("{}", uri))
                            .unwrap()
                            .with_shallow_clone(true)
                    };

                    debug!("About to do a clone!");

                    // clone the repo
                    // Fetch the HEADS refs for all branches
                    let new_repo_ref = repo_watcher.update_state().await.unwrap();

                    // Copy last_known_heads for update in place
                    let mut current_heads: Vec<GitRepoRemoteBranchHead> = last_known_heads.to_vec();

                    // TODO: We may want to process this in parallel (for busy repos)
                    // Loop through the newest HEADS and compare to the DB last known
                    for (branch, head_meta) in new_repo_ref.branch_heads {
                        // Compare the newest branch with the last known

                        //// Copy last_known_heads for update in place
                        //let mut current_heads = last_known_heads.clone();

                        // Dig through `last_known_heads` generated code
                        let mut last_iter = last_known_heads.iter().filter(|b| b.branch == branch);

                        // Compare the DB ref to the new ref, `branch`
                        let _ = if let Some(b) = last_iter.next() {
                            debug!("DB value of branch: {:?}", b);

                            debug!("Value of new commit as String: {:?}", &head_meta.id);

                            //assert!(String::from_utf8(head_meta.id.clone()).is_err());
                            // FIXME: This conversion doesn't yet work
                            let new_commit = head_meta.id.clone();

                            if b.commit != new_commit {
                                info!("There are new commits. Start a new build");

                                // Start a new build
                                // Update DB with new branch reference
                                // Update in DB the current head commit of branch

                                // Find the branch in current_heads, and update the value to the current commit
                                current_heads.iter_mut().for_each(|x| {
                                    if x.branch == branch {
                                        x.commit = new_commit.clone()
                                    }
                                });

                                let db_update_payload = GitRepoUpdateRequest {
                                    org: r.org.clone(),
                                    git_provider: r.git_provider.clone(),
                                    name: r.name.clone(),
                                    user: r.user.clone(),
                                    uri: r.uri.clone(),
                                    canonical_branch: r.canonical_branch.clone(),
                                    secret_type: r.secret_type,
                                    build: r.build,
                                    notify: r.notify,
                                    auth_data: r.auth_data.clone(),
                                    remote_branch_heads: Some(GitRepoRemoteBranchHeadList {
                                        remote_branch_heads: current_heads.clone(),
                                    }),
                                };

                                let _repo_update_res = code_client
                                    .git_repo_update(Request::new(db_update_payload))
                                    .await
                                    .unwrap()
                                    .into_inner();

                                // Skip triggering a build if commit message contains [skip ci] or [ci skip]
                                if is_skip_ci_commit(&head_meta) {
                                    continue;
                                }

                                // Start a new build
                                let build_request = Request::new(BuildTarget {
                                    org: r.org.clone(),
                                    git_repo: r.name.clone(),
                                    remote_uri: r.uri.clone(),
                                    branch: branch.clone(),
                                    commit_hash: new_commit,
                                    //user_envs: None,
                                    trigger: JobTrigger::Poll.into(),
                                    //config: ,
                                    ..Default::default()
                                });

                                info!("Starting build on existing branch from poll!");

                                let mut build_client = BuildServiceClient::connect(format!(
                                    "http://{}",
                                    ORB_DEFAULT_URI
                                ))
                                .await
                                .unwrap();

                                build_client
                                    .build_start(build_request)
                                    .await
                                    .expect("Unable to start build");
                            } else {
                                debug!("No new commits on branch: {}", &branch)
                            }
                        } else {
                            // If not found, then `branch` is new
                            info!("Branch '{}' is new. Start a new build", &branch);

                            info!("Updating the db with new branch reference");

                            let new_branch = GitRepoRemoteBranchHead {
                                branch: branch.clone(),
                                commit: hex::encode(&head_meta.id),
                            };

                            current_heads.push(new_branch);

                            // Update in DB the current head commit of branch
                            let db_update_payload = GitRepoUpdateRequest {
                                org: r.org.clone(),
                                git_provider: r.git_provider.clone(),
                                name: r.name.clone(),
                                user: r.user.clone(),
                                uri: r.uri.clone(),
                                canonical_branch: r.canonical_branch.clone(),
                                secret_type: r.secret_type,
                                build: r.build,
                                notify: r.notify,
                                auth_data: r.auth_data.clone(),
                                remote_branch_heads: Some(GitRepoRemoteBranchHeadList {
                                    remote_branch_heads: current_heads.clone(),
                                }),
                            };

                            let repo_update_res = code_client
                                .git_repo_update(Request::new(db_update_payload))
                                .await
                                .unwrap()
                                .into_inner();

                            debug!(
                                "Results from updating new branch ref: {:?}",
                                repo_update_res
                            );

                            // Skip triggering a build if commit message contains [skip ci] or [ci skip]
                            if is_skip_ci_commit(&head_meta) {
                                continue;
                            }

                            // Start a new build
                            let build_request = Request::new(BuildTarget {
                                org: r.org.clone(),
                                git_repo: r.name.clone(),
                                remote_uri: r.uri.clone(),
                                branch: branch.clone(),
                                commit_hash: hex::encode(head_meta.id.clone()),
                                //user_envs: None,
                                trigger: JobTrigger::Poll.into(),
                                //config: ,
                                ..Default::default()
                            });

                            info!("Starting build on new branch from poll!");

                            let mut build_client =
                                BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI))
                                    .await
                                    .unwrap();

                            build_client
                                .build_start(build_request)
                                .await
                                .expect("Unable to start build");
                        };
                    } // End of branch loop
                } // End of Repos loop
            } // End of Orgs loop

            // TODO: How to load up an ssh key before cloning using the `git` cli?

            //loop {
            //    println!("Poll");
            //    time::delay_for(Duration::from_secs(1)).await;
            //}
            // Wait a few seconds before restarting loop
            time::sleep(Duration::from_secs(poll_freq.into())).await;
        } // End of loop
    });
}

fn is_skip_ci_commit(head_meta: &git_meta::GitCommitMeta) -> bool {
    if let Some(commit_msg) = head_meta.message.clone() {
        if commit_msg.contains("[skip ci]") || commit_msg.contains("[ci skip]") {
            info!("Skipping build due to commit message");
            return true;
        }
    }
    false
}
