extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

// Secrets maybe want to move to credentials?
use orbital_headers::credential::{client::CredentialServiceClient, SecretCreateRequest};

use crate::ORB_DEFAULT_URI;
use tonic::Request;

/// Local options for customizing secrets request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

// FIXME:
/// *Not yet implemented* Backend calls for managing secrets resources
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client =
        CredentialServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(SecretCreateRequest {
        org: "org_name_goes_here".into(),
        secret_type: 0,
        secret: "secret_text_goes_here".into(),
    });

    let response = client.create_secret(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}
