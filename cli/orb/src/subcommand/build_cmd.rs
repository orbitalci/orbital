extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper;
use tower_util::MakeService;

/// Local options for customizing build start request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// Generates gRPC `BuildStartRequest` object and connects to *currently hardcoded* gRPC server and sends a request to `BuildService` server.
pub fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let uri: http::Uri = format!("http://127.0.0.1:50051").parse().unwrap();

    let dst = Destination::try_from_uri(uri.clone()).unwrap();
    let connector = tower_hyper::util::Connector::new(HttpConnector::new(4));
    let settings = tower_hyper::client::Builder::new().http2_only(true).clone();
    let mut make_client = tower_hyper::client::Connect::with_builder(connector, settings);

    let say_hello = make_client
        .make_service(dst)
        .map_err(|e| panic!("connect error: {:?}", e))
        .and_then(move |conn| {
            use orbital_headers::build_metadata::client::BuildService;

            let conn = tower_request_modifier::Builder::new()
                .set_origin(uri)
                .build(conn)
                .unwrap();

            // Wait until the client is ready...
            BuildService::new(conn).ready()
        })
        .and_then(|mut client| {
            use orbital_headers::build_metadata::BuildStartRequest;

            client.start_build(Request::new(BuildStartRequest {
                remote_uri: "What is in a name?".to_string(),
                branch: "What is in a name?".to_string(),
                commit_ref: "What is in a name?".to_string(),
            }))
        })
        .and_then(|response| {
            println!("RESPONSE = {:?}", response);
            Ok(())
        })
        .map_err(|e| {
            println!("ERR = {:?}", e);
        });

    tokio::run(say_hello);

    Ok(())
}
