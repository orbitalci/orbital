use structopt::StructOpt;

use crate::subcommand::{secret::SubcommandOption, GlobalOption};

use crate::orbital_database::postgres::schema::SecretType;
use crate::orbital_headers::secret::{
    secret_service_client::SecretServiceClient, SecretGetRequest,
};
use crate::orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use crate::orbital_database::postgres::secret::Secret;
use color_eyre::eyre::Result;
use prettytable::{cell, format, row, Table};
use strum::VariantNames;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Secret name
    #[structopt(required = true)]
    secret_name: String,

    /// Secret Type
    #[structopt(long, required = true, possible_values = &SecretType::VARIANTS)]
    secret_type: SecretType,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    let mut client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(SecretGetRequest {
        org: action_option.org.unwrap_or_default().into(),
        name: action_option.secret_name.into(),
        secret_type: action_option.secret_type.into(),
    });

    debug!("Request for secret get: {:?}", &request);

    let response = client.secret_get(request).await;

    match response {
        Err(_e) => {
            eprintln!("Secret not found");
            Ok(())
        }
        Ok(s) => {
            let secret_proto = s.into_inner();

            debug!("RESPONSE = {:?}", &secret_proto);

            // By default, format the response into a table
            let mut table = Table::new();
            table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

            // Print the header row
            table.set_titles(row![
                bc =>
                "Org Name",
                "Secret Name",
                "Secret Type",
                "Vault Path",
                "Active State",
            ]);

            let secret = Secret::from(secret_proto.clone());

            table.add_row(row![
                secret_proto.org,
                secret.name,
                &format!("{:?}", secret.secret_type),
                secret.vault_path,
                &format!("{:?}", secret.active_state),
            ]);

            // Print the table to stdout
            table.printstd();
            Ok(())
        }
    }
}
