//ocelot creds ssh add --identifier JESSI_SSH_KEY --acctname level11consulting --sshfile-loc /Users/jesseshank/.ssh/id_rsa

extern crate structopt;
use structopt::StructOpt;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

use git_meta::git_info;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Identifier
    #[structopt(name = "Identifier", long, alias = "id")]
    identifier: String,
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
    /// File path to SSH Private key (type RSA-only)
    #[structopt(name = "SSH rsa private key", short = "f", long = "file")]
    file_path: Option<String>,
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
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
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
                    use ocelot_api::protobuf_api::legacyapi::SshKeyWrapper;
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;

                    let mut ssh_key_proto = SshKeyWrapper::default();

                    ssh_key_proto.acct_name = account;
                    ssh_key_proto.sub_type = SubCredType::Sshkey.into();
                    ssh_key_proto.identifier = identifier.to_string();
                    ssh_key_proto.private_key = file_contents.into_bytes();

                    // Send off a build info request
                    // Only supports bitbucket right now
                    client.set_ssh_creds(Request::new(ssh_key_proto))
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
        ResourceAction::Delete(args) => {
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
                    use ocelot_api::protobuf_api::legacyapi::SshKeyWrapper;
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;

                    let mut ssh_key_proto = SshKeyWrapper::default();

                    ssh_key_proto.acct_name = account;
                    ssh_key_proto.sub_type = SubCredType::Sshkey.into();
                    ssh_key_proto.identifier = identifier.to_string();

                    // Send off a build info request
                    // Only supports bitbucket right now
                    client.delete_ssh_creds(Request::new(ssh_key_proto))
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
        ResourceAction::List(_args) => {
            // We don't seem to pass this info in our request
            //let path_to_repo = ocelot_api::client_util::get_repo(args.path.clone());
            //let git_info = git_info::get_git_info_from_path(&path_to_repo, &None, &None);
            //let account = args.account.unwrap_or(git_info.account);

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
                .and_then(move |mut client| client.get_ssh_creds(Request::new(())))
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
