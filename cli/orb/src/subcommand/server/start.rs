use anyhow::Result;
use structopt::StructOpt;

use crate::GlobalOption;

use orbital_headers::{
    build_meta::build_service_server::BuildServiceServer,
    code::code_service_server::CodeServiceServer,
    notify::notify_service_server::NotifyServiceServer,
    organization::organization_service_server::OrganizationServiceServer,
    secret::secret_service_server::SecretServiceServer,
};
use orbital_services::OrbitalApi;

use orbital_services::ORB_DEFAULT_URI;

use log::debug;
use std::path::PathBuf;
use tonic::transport::Server;

/// Local options for starting build service
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,
}

/// Binds a *currently hardcoded* address and starts all services on mutliplexed gRPC server
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<()> {
    let addr = ORB_DEFAULT_URI.parse().unwrap();

    debug!("Starting single-node server");
    Server::builder()
        .add_service(BuildServiceServer::new(OrbitalApi::default()))
        .add_service(CodeServiceServer::new(OrbitalApi::default()))
        .add_service(NotifyServiceServer::new(OrbitalApi::default()))
        .add_service(OrganizationServiceServer::new(OrbitalApi::default()))
        .add_service(SecretServiceServer::new(OrbitalApi::default()))
        .serve(addr)
        .await?;

    Ok(())
}
