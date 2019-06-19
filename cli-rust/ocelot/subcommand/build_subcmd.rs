/// This is named build_subcmd.rs bc we can't use build.rs due to overlapping with `cargo` features.

use structopt::StructOpt;

use std::env;

use ocelot_api;
use git_meta::git_info;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Build provided account/repo. Otherwise try to auto-detect from current working directory
    #[structopt(long)]
    acct_repo: Option<String>,
    /// Use provided local branch. Default to current active branch
    #[structopt(long)]
    branch: Option<String>,
    /// Build provided commit hash. Otherwise, default to HEAD commit of active branch
    #[structopt(long)]
    hash: Option<String>,
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
    
}

// The goal of this command
// If we pass a commit hash alone, we assume the current branch.
//      If no, then we might end up in a detached HEAD? Try to find if commit is in the working branch and use that. Otherwise checkout detached HEAD
//      If yes, then we should pass back a remote ref to the branch+commit

// If we pass a local branch alone, we should resolve the branch to a remote ref HEAD

// Passing both the branch and commit should resolve to that specific remote ref

// TODO: Return a Result for the questionmark operator
// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {

    // Parse the following information from the local repo before calling the backend
    //
    // git provider Account name
    // Repo (something.git)
    // Provider (like bitbucket or github account)
    // Remote Branch
    // Target commit, or HEAD of the remote branch if not specified
    // Later: Env vars

    // Assume current directory for now
    let path_to_repo = args.path.clone()
                        .unwrap_or(
                            env::current_dir()
                            .unwrap().to_str()
                            .unwrap()
                            .to_string()
                        );

    println!("Path to repo: {:?}", path_to_repo);

    // Get the git info from the path
    let git_info = git_info::get_git_info_from_path(&path_to_repo, &args.branch, &args.hash);
    println!("Git info: {:?}",git_info);

    // TODO: Factor this out later
    // Connect to Ocelot server via grpc.
    let uri: http::Uri = format!("http://192.168.12.34:10000").parse().unwrap();
    let dst = Destination::try_from_uri(uri.clone()).unwrap();

    let connector = util::Connector::new(HttpConnector::new(4));
    let settings = client::Builder::new().http2_only(true).clone();
    let mut make_client = client::Connect::with_builder(connector, settings);

    let build_req = make_client
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
        .and_then(|mut client| {
            use ocelot_api::protobuf_api::legacyapi::BuildReq;

            // Send off a build info request
            // Only supports bitbucket right now
            client.build_repo_and_hash(Request::new(BuildReq {
                acct_repo: format!("{}/{}",git_info.account,git_info.repo),
                hash: git_info.id,
                branch: git_info.branch,
                force: false,
                vcs_type: 1,
            }))
        })
        .and_then(|response| {
            println!("RESPONSE = {:?}", response);
            Ok(())
        })
        .map_err(|e| {
            println!("ERR = {:?}", e);
        });

    tokio::run(build_req);
}
