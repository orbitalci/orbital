extern crate structopt;
use structopt::StructOpt;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,

    /// Use the provided acct-repo
    #[structopt(long)]
    acct_repo: Option<String>,

    /// Comma-separated list of branches
    #[structopt(alias = "branches")]
    branch: Option<String>,
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
    /// Register a git repo
    Add(AddOption),
    /// Delete a registered git repo
    #[structopt(alias = "rm")]
    Delete(DeleteOption),
    /// List the registered git repo(s)
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
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling Git repos");

    match &args.action {
        ResourceAction::Add(_) => println!("Note: There is no GRPC endpoint to add repos"),
        ResourceAction::Delete(_) => println!("Note: There is no GRPC endpoint to delete repos"),
        ResourceAction::List(_) => {
            // TODO: Factor this out later
            // Connect to Ocelot server via grpc.
            let uri: http::Uri = format!("http://192.168.12.34:10000").parse().unwrap();
            let dst = Destination::try_from_uri(uri.clone()).unwrap();

            let connector = util::Connector::new(HttpConnector::new(4));
            let settings = client::Builder::new().http2_only(true).clone();
            let mut make_client = client::Connect::with_builder(connector, settings);

            let repo_req = make_client
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
                    client.get_tracked_repos(Request::new(()))
                })
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(repo_req);
        }
    }
}
