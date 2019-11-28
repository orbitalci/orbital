use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::secret::{client::SecretServiceClient, SecretAddRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use hashicorp_stack::vault;

use log::debug;

use std::fs::File;
use std::io::prelude::*;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    _action_option: ActionOption,
) -> Result<(), SubcommandError> {

    let mut file = File::open("/home/telant/.ssh/id_ed25519")?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;

    let _ = vault::vault_add_secret(
        "orbital/default_org/github.com/level11consulting/orbital",
        contents.as_str(),
    );

    let mut client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(SecretAddRequest {
        ..Default::default()
    });

    debug!("Request for secret add: {:?}", &request);

    let response = client.secret_add(request).await?;
    println!("RESPONSE = {:?}", response);

    Ok(())
}
