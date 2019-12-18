use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::orbital_types::SecretType;
use orbital_headers::secret::{client::SecretServiceClient, SecretUpdateRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use std::fs::File;
use std::io::prelude::*;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Secret name
    #[structopt(required = true)]
    secret_name: String,

    /// Secret Type
    #[structopt(long, required = true, possible_values = &SecretType::variants())]
    secret_type: SecretType,

    /// Secret filepath
    #[structopt(long, short = "f")]
    secret_file: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut file = File::open(&action_option.secret_file.expect("No secret filepath given"))?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;

    let mut client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(SecretUpdateRequest {
        org: action_option.org.unwrap_or_default(),
        name: action_option.secret_name.into(),
        secret_type: action_option.secret_type.into(),
        data: contents.into(),
        ..Default::default()
    });

    debug!("Request for secret update: {:?}", &request);

    let response = client.secret_update(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}
