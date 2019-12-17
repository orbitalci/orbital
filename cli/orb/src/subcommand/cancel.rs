use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_meta::{client::BuildServiceClient, BuildTarget};

use orbital_services::ORB_DEFAULT_URI;
use std::path::PathBuf;
use tonic::Request;

/// Local options for customizing build cancel request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

/// Generates request for canceling a build in progress
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(BuildTarget {
        id: 0,
        ..Default::default()
    });

    let response = client.build_stop(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}
