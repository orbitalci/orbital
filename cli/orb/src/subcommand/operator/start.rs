extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use futures::{Future, Stream};
use orbital_headers::build_metadata::server::BuildServiceServer;
use orbital_services::OrbitalApi;

use log::error;
use tokio::net::TcpListener;
use tower_hyper::server::{Http, Server};

/// Local options for starting build service
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// Binds a *currently hardcoded* address and starts a `BuildService` gRPC server
pub fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let new_service = BuildServiceServer::new(OrbitalApi);
    let mut server = Server::new(new_service);
    let http = Http::new().http2_only(true).clone();
    let addr = "127.0.0.1:50051".parse().unwrap();
    let bind = TcpListener::bind(&addr).expect("bind");

    let serve = bind
        .incoming()
        .for_each(move |sock| {
            if let Err(e) = sock.set_nodelay(true) {
                return Err(e);
            }

            let serve = server.serve_with(sock, http.clone());
            tokio::spawn(serve.map_err(|e| error!("hyper error: {:?}", e)));

            Ok(())
        })
        .map_err(|e| eprintln!("accept error: {}", e));

    tokio::run(serve);

    Ok(())
}
