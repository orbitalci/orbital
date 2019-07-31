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
    /// File path to config yaml
    #[structopt(name = "Path to config yaml", short = "f", long = "file")]
    file_path: Option<String>,
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct DeleteOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
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
    /// Add a Version Control System
    Add(AddOption),
    /// Delete a Version Control System
    #[structopt(alias = "rm")]
    Delete(DeleteOption),
    /// List registered Version Control System
    #[structopt(alias = "ls")]
    List(ListOption),
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    #[structopt(flatten)]
    action: ResourceAction,

    #[structopt(long)]
    acct: Option<String>,
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

            let file_contents = ocelot_api::client_util::read_file(args.file_path.clone());
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
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;
                    use ocelot_api::protobuf_api::legacyapi::VcsCreds;

                    let mut vcs_creds = VcsCreds::default();

                    // TODO: This needs to be completed
                    vcs_creds.acct_name = account;
                    vcs_creds.sub_type = SubCredType::Bitbucket.into();
                    vcs_creds.identifier = identifier.to_string();
                    vcs_creds.client_id = "".to_string();
                    vcs_creds.client_secret = file_contents.to_string();
                    vcs_creds.token_url = "".to_string();

                    // Send off a build info request
                    // Only supports bitbucket right now
                    client.set_vcs_creds(Request::new(vcs_creds))
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
        ResourceAction::Delete(_args) => println!("There is no grpc endpoint for deleting VCS"),
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
                .and_then(move |mut client| client.get_vcs_creds(Request::new(())))
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
