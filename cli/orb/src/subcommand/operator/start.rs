extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::{
    build_metadata::server::BuildServiceServer, credential::server::CredentialServiceServer,
    integration::server::IntegrationServiceServer, organization::server::OrganizationServiceServer,
};
use orbital_services::OrbitalApi;

use crate::ORB_DEFAULT_URI;

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

/// Binds a *currently hardcoded* address and starts all services on mutliplexed gRPC server
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let addr = ORB_DEFAULT_URI.parse().unwrap();

    debug!("Starting single-node server");
    Server::builder()
        .add_service(BuildServiceServer::new(OrbitalApi::default()))
        .add_service(CredentialServiceServer::new(OrbitalApi::default()))
        .add_service(IntegrationServiceServer::new(OrbitalApi::default()))
        .add_service(OrganizationServiceServer::new(OrbitalApi::default()))
        .serve(addr)
        .await?;

    Ok(())
}
