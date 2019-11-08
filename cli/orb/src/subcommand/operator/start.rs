extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_metadata::server::BuildServiceServer;
use orbital_services::OrbitalApi;

use log::debug;
use tonic::transport::Server;

/// Local options for starting build service
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// Binds a *currently hardcoded* address and starts a `BuildService` gRPC server
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let addr = "127.0.0.1:50051".parse().unwrap();
    let buildserver = OrbitalApi::default();

    debug!("Starting BuildService server");
    Server::builder()
        .add_service(BuildServiceServer::new(buildserver))
        .serve(addr)
        .await?;

    Ok(())
}
