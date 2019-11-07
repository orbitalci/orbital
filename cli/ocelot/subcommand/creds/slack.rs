//ocelot creds notify add --identifier L11_SLACK --acctname level11consulting --url https://hooks.slack.com/services/T0DFsdSBA/345PPRP9C/5hUe12345v6BrxfSJt --detail-url https://ocelot.mysite.io

extern crate structopt;
use structopt::StructOpt;

use git_meta::git_info;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Identifier
    #[structopt(name = "Identifier", long, alias = "id")]
    identifier: String,
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
    /// Kubernetes cluster name
    #[structopt(name = "Slack org name", long)]
    slack_name: Option<String>,
    /// File path to yaml containing env vars
    #[structopt(name = "Kubernetes config (yaml)", short = "f", long = "file")]
    webhook_url: Option<String>,
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct DeleteOption {
    /// Identifier
    #[structopt(name = "Identifier", long, alias = "id")]
    identifier: String,
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
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
pub enum ResourceAction {
    ///
    Add(AddOption),
    ///
    #[structopt(alias = "rm")]
    Delete(DeleteOption),
    ///
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

    match args.action {
        ResourceAction::Add(args) => {
            let path_to_repo = ocelot_api::client_util::get_repo(args.path.clone());
            let git_info = git_info::get_git_info_from_path(&path_to_repo, &None, &None);
            let account = args.account.unwrap_or(git_info.account);

            let identifier = args.identifier.clone();

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
                    use ocelot_api::protobuf_api::legacyapi::NotifyCreds;
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;

                    let mut notify_creds = NotifyCreds::default();

                    // TODO: This needs to be completed
                    notify_creds.acct_name = account;
                    notify_creds.sub_type = SubCredType::Slack.into();
                    notify_creds.identifier = identifier.to_string();
                    notify_creds.client_secret = "".to_string();
                    notify_creds.detail_url_base = "".to_string();

                    // Send off a build info request
                    // Only supports bitbucket right now
                    client.set_notify_creds(Request::new(notify_creds))
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
        ResourceAction::Delete(_args) => {println!("There is no grpc endpoint for deleting Slack webhook")},
        ResourceAction::List(_args) => {
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
                .and_then(move |mut client| client.get_notify_creds(Request::new(())))
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
