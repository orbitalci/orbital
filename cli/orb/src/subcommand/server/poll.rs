use log::{debug, info};
use tokio::time::{self, Duration};

use git_event;
use orbital_headers::code::{
    code_service_client::CodeServiceClient, GitRepoEntry, GitRepoListRequest, GitRepoListResponse,
};
use orbital_headers::organization::organization_service_client::OrganizationServiceClient;
use orbital_services::ORB_DEFAULT_URI;
use std::str;
use tonic::Request;
use hex;

pub async fn poll_for_new_commits() {
    tokio::spawn(async move {
        println!("Waiting a moment for server to start before polling");

        loop {
            time::delay_for(Duration::from_secs(30)).await;
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

            let mut repos_response: Vec<GitRepoEntry> = Vec::new();

            info!("Looping through orgs to collect their repos");
            for org in orgs {
                let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI))
                    .await
                    .unwrap();
                let request = Request::new(GitRepoListRequest {
                    org: org.clone(),
                    ..Default::default()
                });

                info!("Collecting repos for org: {}", &org);
                repos_response = client
                    .git_repo_list(request)
                    .await
                    .unwrap()
                    .into_inner()
                    .git_repos;

                debug!("repos: {:?}", repos_response);

                // Disconnect from CodeService
                drop(client);

                // FIXME: If auth required, gather auth from SecretService

                // Load up a list of all the repos, and the path to any secrets for cloning

                // Per branch, compare the last known remote HEAD commit, and the current remote commit.

                for r in repos_response {
                    // NEED: Last known commit from DB
                    // FIXME: Yikes.
                    let last_known_heads = r
                        .remote_branch_head_refs
                        .expect("No last known commits in DB")
                        .remote_branch_head_refs;

                    let repo_watcher = git_event::GitRepoWatchHandler::new(&r.uri)
                        .unwrap()
                        .with_shallow_clone(true);

                    // clone the repo
                    // Fetch the HEADS refs for all branches
                    let new_repo_ref = repo_watcher.oneshot_report().await.unwrap();

                    // Loop through the newest HEADS and compare to the DB last known
                    for (branch, head_meta) in new_repo_ref.branch_heads {
                        // Compare the newest branch with the last known

                        // Dig through `last_known_heads` generated code
                        let mut last_iter = last_known_heads.iter().filter(|b| b.branch == branch);

                        // Compare the DB ref to the new ref, `branch`
                        let _ = if let Some(b) = last_iter.next() {
                            println!("DB value of branch: {:?}", b);

                            println!("Value of new commit as String: {:?}", hex::encode(&head_meta.id));

                            //assert!(String::from_utf8(head_meta.id.clone()).is_err());
                            // FIXME: This conversion doesn't yet work
                            let new_commit = hex::encode(head_meta.id);


                            if b.commit != new_commit {
                                info!("There are new commits. Start a new build")

                            // Start a new build
                            // Update DB with new branch reference
                            } else {
                                debug!("Commits are the same")
                            }
                        } else {
                            // If not found, then `branch` is new
                            info!("Branch '{}' is new. Start a new build", &branch)

                            // Start a new build
                            // Update DB with new branch reference
                        };
                    } // End of branch loop
                } // End of Repos loop
            } // End of Orgs loop

            // TODO: How to load up an ssh key before cloning using the `git` cli?

            //loop {
            //    println!("Poll");
            //    time::delay_for(Duration::from_secs(1)).await;
            //}
        } // End of loop
    });
}
