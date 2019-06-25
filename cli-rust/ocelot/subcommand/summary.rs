extern crate structopt;
use structopt::StructOpt;

use std::env;

use git_meta::git_info;
use ocelot_api;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    // build-id will provide the same functionality that the `status` subcommand did.
    /// Retrieve status for specific build
    #[structopt(name = "build id", long)]
    build_id: Option<u32>,
    /// Retrieve status for builds from the provided acct-repo
    #[structopt(long)]
    acct_repo: Option<String>,
    /// Retrieve status for builds from the provided branch.
    //#[structopt(long)]
    //branch : Option<String>,
    /// Retrieve status for builds with the provided commit hash
    #[structopt(long)]
    hash: Option<String>,
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
    /// Limit to last N runs
    #[structopt(long)]
    limit: Option<i32>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    // Assume current directory for now
    let path_to_repo = args
        .path
        .clone()
        .unwrap_or(env::current_dir().unwrap().to_str().unwrap().to_string());

    println!("Path to repo: {:?}", path_to_repo);

    // Get the git info from the path
    let git_info = git_info::get_git_info_from_path(&path_to_repo, &None, &args.hash);
    println!("Git info: {:?}", git_info);

    let results_limit = args.limit.unwrap_or(10);

    // TODO: Factor this out later
    // Connect to Ocelot server via grpc.
    let uri: http::Uri = format!("http://192.168.12.34:10000").parse().unwrap();
    let dst = Destination::try_from_uri(uri.clone()).unwrap();

    let connector = util::Connector::new(HttpConnector::new(4));
    let settings = client::Builder::new().http2_only(true).clone();
    let mut make_client = client::Connect::with_builder(connector, settings);

    let summary_req = make_client
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
            use ocelot_api::protobuf_api::legacyapi::RepoAccount;

            // Send off a build info request
            // Only supports bitbucket right now
            client.last_few_summaries(Request::new(RepoAccount {
                repo: git_info.repo,
                account: git_info.account,
                limit: results_limit,
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

    tokio::run(summary_req);
}
