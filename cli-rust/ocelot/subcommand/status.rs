use structopt::StructOpt;

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
    /// Status for provided account. Otherwise try to auto-detect from current working directory
    #[structopt(long)]
    account: Option<String>,
    /// Status for provided repo. Otherwise try to auto-detect from current working directory
    #[structopt(long)]
    repo: Option<String>,
    /// Status for  provided commit hash. Otherwise, default to HEAD commit of active branch
    #[structopt(long)]
    hash: Option<String>,
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

pub fn subcommand_handler(args: SubOption) {
    let uri = ocelot_api::client_util::get_client_uri();
    let dst = Destination::try_from_uri(uri.clone()).unwrap();

    let connector = util::Connector::new(HttpConnector::new(4));
    let settings = client::Builder::new().http2_only(true).clone();
    let mut make_client = client::Connect::with_builder(connector, settings);

    let path_to_repo = ocelot_api::client_util::get_repo(args.path.clone());
    let git_info = git_info::get_git_info_from_path(&path_to_repo, &None, &None);
    let account = args.account.unwrap_or(git_info.account.clone());

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
        .and_then(|mut client| {
            use ocelot_api::protobuf_api::legacyapi::StatusQuery;

            let mut status_query = StatusQuery::default();
            status_query.acct_name = account;
            status_query.repo_name = git_info.repo;

            // Send off a build info request
            // Only supports bitbucket right now
            client.get_status(Request::new(status_query))
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
