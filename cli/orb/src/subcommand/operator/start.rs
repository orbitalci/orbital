extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use futures::{Future, Stream};
use orbital_services::build_service;

use log::error;
use tokio::net::TcpListener;
use tower_hyper::server::{Http, Server};

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

pub fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    //let handler : build_service::OrbitalApi;
    let new_service =
        orbital_headers::builder::server::BuildServiceServer::new(build_service::OrbitalApi);
    let mut server = Server::new(new_service);
    let http = Http::new().http2_only(true).clone();
    let addr = "[::1]:50051".parse().unwrap();
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
