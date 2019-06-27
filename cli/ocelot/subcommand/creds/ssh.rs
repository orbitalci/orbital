//ocelot creds ssh add --identifier JESSI_SSH_KEY --acctname level11consulting --sshfile-loc /Users/jesseshank/.ssh/id_rsa

extern crate structopt;
use structopt::StructOpt;
use std::env;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

use std::fs::File;
use std::io::prelude::*;

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
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling SSH creds");

    match &args.action {
        ResourceAction::Add(args) => {
            use git_meta::git_info;

            let identifier = args.identifier.clone();

            // Assume current directory for now
            let path_to_repo = &args
                .path
                .clone()
                .unwrap_or(env::current_dir().unwrap().to_str().unwrap().to_string());

            let mut file = File::open(&args.file_path.clone().unwrap()).unwrap();
            let mut contents = String::new();
            file.read_to_string(&mut contents).unwrap();

            println!("Path to repo: {:?}", path_to_repo);

            // Get the git info from the path
            let git_info = git_info::get_git_info_from_path(&path_to_repo, &None, &None);
            println!("Git info: {:?}", git_info);

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
                    use ocelot_api::protobuf_api::legacyapi::SshKeyWrapper;
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;

                    let mut ssh_key_proto = SshKeyWrapper::default();

                    ssh_key_proto.acct_name = git_info.account;
                    ssh_key_proto.sub_type = SubCredType::Sshkey.into();
                    ssh_key_proto.identifier = identifier.to_string();
                    ssh_key_proto.private_key = contents.into_bytes();

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

            tokio::run(repo_req);
        },
        ResourceAction::Delete(args) => {
            use git_meta::git_info;

            let identifier = args.identifier.clone();

            // Assume current directory for now
            let path_to_repo = &args
                .path
                .clone()
                .unwrap_or(env::current_dir().unwrap().to_str().unwrap().to_string());

            // Get the git info from the path
            let git_info = git_info::get_git_info_from_path(&path_to_repo, &None, &None);
            println!("Git info: {:?}", git_info);

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
                    use ocelot_api::protobuf_api::legacyapi::SshKeyWrapper;
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;

                    let mut ssh_key_proto = SshKeyWrapper::default();

                    ssh_key_proto.acct_name = git_info.account;
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

            tokio::run(repo_req);
        },
        ResourceAction::List(args) => {
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
                    client.get_ssh_creds(Request::new(()))
                })
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(repo_req);
        },
    }
}
