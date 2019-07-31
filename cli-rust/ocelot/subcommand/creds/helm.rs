//ocelot creds helmrepo add -acct my_kewl_acct -repo-name shankj3_charts -helm-url https://github.io/shankj3_helm_repository
//ocelot creds helmrepo list -account <ACCT_NAME>

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
    account: String,
    /// Helm repo name (logical)
    #[structopt(name = "Helm repo name", long)]
    helm_name: String,
    /// Helm repo url
    #[structopt(name = "Helm repo url", long)]
    helm_url: String,
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
            let account = args.account;
            let identifier = args.helm_name;
            let secret = args.helm_url;

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
                    use ocelot_api::protobuf_api::legacyapi::GenericCreds;
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;

                    let mut env_proto = GenericCreds::default();

                    env_proto.acct_name = account;
                    env_proto.sub_type = SubCredType::HelmRepo.into();
                    env_proto.identifier = identifier;
                    env_proto.client_secret = secret;

                    client.set_generic_creds(Request::new(env_proto))
                })
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(req)
        }
        ResourceAction::Delete(_args) => {
            println!("There is no grpc endpoint for deleting helm repos")
        }
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
                .and_then(move |mut client| client.get_generic_creds(Request::new(())))
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(req)
        }
    }
}
