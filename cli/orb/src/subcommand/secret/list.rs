use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::secret::{secret_service_client::SecretServiceClient, SecretListRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use orbital_database::postgres::secret::Secret;
use prettytable::{cell, format, row, Table};
use orbital_headers::orbital_types;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
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

    let request = Request::new(SecretListRequest {
        org: action_option.org.unwrap_or_default(),
        ..Default::default()
    });

    debug!("Request for secret list: {:?}", &request);

    let response = client.secret_list(request).await?.into_inner();

    // By default, format the response into a table
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

    // Print the header row
    table.set_titles(row![
        "Org Name",
        "Secret Name",
        "Secret Type",
        "Vault Path",
        "Active State",
    ]);

    match response.secret_entries.len() {
        0 => {
            println!("No secrets found");
        }
        _ => {
            for secret_proto in &response.secret_entries {
                let secret = Secret::from(secret_proto.clone());

                table.add_row(row![
                    secret_proto.org,
                    secret.name,
                    &format!("{:?}", secret.secret_type),
                    secret.vault_path,
                    &format!("{:?}", secret.active_state),
                ]);
            }

            debug!("RESPONSE = {:?}", &response);
        }
    }

    // Print the table to stdout
    table.printstd();

    Ok(())
}
