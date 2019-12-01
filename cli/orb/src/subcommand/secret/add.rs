use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::orbital_types::SecretType;
use orbital_headers::secret::{client::SecretServiceClient, SecretAddRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use std::fs::File;
use std::io::prelude::*;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Secret name
    #[structopt(required = true)]
    secret_name: String,

    /// Secret Type
    #[structopt(long, required = true, possible_values = &[
        "basicauth", 
        "apikey", 
        "envvar", 
        "file", 
        "sshkey", 
        "dockerregistry", 
        "npmrepo", 
        "pypiregistry", 
        "mavenrepo", 
        "kubernetes"])]
    secret_type: String,

    #[structopt(long, default_value = "default_org")]
    org_name: String,

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

    let org_name = &action_option.org_name;
    let secret_name = &action_option.secret_name;
    let vault_path = format!(
        "orbital/{}/{}/{}",
        org_name,
        SecretType::from(action_option.secret_type.clone()),
        secret_name,
    )
    .to_lowercase();

    let request = Request::new(SecretAddRequest {
        org: org_name.into(),
        name: vault_path.into(),
        secret_type: SecretType::from(action_option.secret_type.clone()).into(),
        data: contents.into(),
    });

    debug!("Request for secret add: {:?}", &request);

    let response = client.secret_add(request).await?;
    println!("RESPONSE = {:?}", response);

    Ok(())
}
