use tokio::time::{self, Duration};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoListRequest};
use orbital_headers::organization::organization_service_client::OrganizationServiceClient;
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

pub async fn poll_for_new_commits() {
    tokio::spawn(async move {
        println!("Waiting a moment for server to start before polling");
        // Wait for the server to be available
        time::delay_for(Duration::from_secs(10)).await;

        // Gather all orgs

        let mut client = OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI))
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

        println!("Orgs: {:?}", orgs);

        // Gather all repos per org

        let mut repos;

        for org in orgs {
            let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI))
                .await
                .unwrap();
            let request = Request::new(GitRepoListRequest {
                org: org,
                ..Default::default()
            });

            repos = client
                .git_repo_list(request)
                .await
                .unwrap()
                .into_inner()
                .git_repos;

            println!("repos: {:?}", repos);
        }

        // Load up a list of all the repos, and the path to any secrets for cloning

        // Per branch, compare the last known remote HEAD commit, and the current remote commit.

        // TODO: How to load up an ssh key before cloning using the `git` cli?

        //loop {
        //    println!("Poll");
        //    time::delay_for(Duration::from_secs(1)).await;
        //}
    });
}
