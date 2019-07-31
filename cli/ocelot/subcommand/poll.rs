//  ocelot poll delete -acct-repo level11consulting/ocelog
//      -- I'm not sure we should force the --acct-repo flag
// Should have add as a subcommand instead of the current functionality

// ocelot poll list might want to filter by account

extern crate structopt;
use std::env;
use structopt::StructOpt;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

use std::collections::HashMap;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,

    /// Use the provided repo
    #[structopt(long)]
    repo: Option<String>,

    /// Cron string
    #[structopt(long = "cron")]
    cron_string: Option<String>,

    /// Comma-separated list of branches
    #[structopt(alias = "branches")]
    branch: Option<String>,

    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct ListOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct DeleteOption {
    /// Delete the poll schedule for the provided account/repo
    #[structopt(long)]
    acct_repo: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    /// Add a polling schedule
    Add(AddOption),
    /// Delete a polling schedule
    #[structopt(alias = "rm")]
    Delete(DeleteOption),
    /// List the polling schedules
    #[structopt(alias = "ls")]
    List(ListOption),
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    #[structopt(flatten)]
    action: ResourceAction,

    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: SubOption) {
    let uri = ocelot_api::client_util::get_client_uri();
    let dst = Destination::try_from_uri(uri.clone()).unwrap();

    let connector = util::Connector::new(HttpConnector::new(4));
    let settings = client::Builder::new().http2_only(true).clone();
    let mut make_client = client::Connect::with_builder(connector, settings);

    match &args.action {
        ResourceAction::Add(args) => {
            use git_meta::git_info;

            // Assume current directory for now
            let path_to_repo = args
                .path
                .clone()
                .unwrap_or(env::current_dir().unwrap().to_str().unwrap().to_string());

            println!("Path to repo: {:?}", path_to_repo);

            // Get the git info from the path
            let git_info = git_info::get_git_info_from_path(&path_to_repo, &args.branch, &None);
            println!("Git info: {:?}", git_info);

            let req = make_client
                .make_service(dst)
                .map_err(|e| panic!("connect error: {:?}", e))
                .and_then(move |conn| {
                    use ocelot_api::protobuf_api::legacyapi::client;

                    let conn = tower_request_modifier::Builder::new()
                        .set_origin(uri)
                        .build(conn)
                        .unwrap();

                    // Wait until the client is ready...
                    client::GuideOcelot::new(conn).ready()
                })
                .and_then(move |mut client| {
                    use ocelot_api::protobuf_api::legacyapi::PollRequest;

                    // Send off a build info request
                    // Only supports bitbucket right now
                    client.poll_repo(Request::new(PollRequest {
                        account: git_info.account,
                        repo: git_info.repo,
                        branches: git_info.branch,
                        cron: "* * * * *".to_string(),
                        last_cron_time: None,
                        last_hashes: HashMap::new(),
                        r#type: 1,
                    }))
                })
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(req);
        }
        ResourceAction::Delete(_) => println!("Note: There is no GRPC endpoint to delete repos"),
        ResourceAction::List(_) => {
            let req = make_client
                .make_service(dst)
                .map_err(|e| panic!("connect error: {:?}", e))
                .and_then(move |conn| {
                    use ocelot_api::protobuf_api::legacyapi::client;

                    let conn = tower_request_modifier::Builder::new()
                        .set_origin(uri)
                        .build(conn)
                        .unwrap();

                    // Wait until the client is ready...
                    client::GuideOcelot::new(conn).ready()
                })
                .and_then(move |mut client| {
                    // Send off a build info request
                    // Only supports bitbucket right now
                    client.list_polled_repos(Request::new(()))
                })
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(req);
        }
    }
}
