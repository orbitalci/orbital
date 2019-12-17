use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::orbital_types::SecretType;
use orbital_headers::secret::{client::SecretServiceClient, SecretGetRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Secret name
    #[structopt(required = true)]
    secret_name: String,

    /// Secret Type
    #[structopt(long, required = true, possible_values = &SecretType::variants())]
    secret_type: SecretType,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let org_name = action_option.org.unwrap_or_default();
    let secret_name = &action_option.secret_name;

    let vault_path = format!(
        "orbital/{}/{}/{}",
        org_name,
        action_option.secret_type.to_string(),
        secret_name
    )
    .to_lowercase();

    let request = Request::new(SecretGetRequest {
        org: org_name.into(),
        name: vault_path,
        secret_type: action_option.secret_type.into(),
    });

    debug!("Request for secret get: {:?}", &request);

    let response = client.secret_get(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}
