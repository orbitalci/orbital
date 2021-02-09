use structopt::StructOpt;

use crate::{org::SubcommandOption, GlobalOption};

use orbital_headers::organization::{
    organization_service_client::OrganizationServiceClient, OrgRemoveRequest,
};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use prettytable::{cell, format, row, Table};

use color_eyre::eyre::Result;
use orbital_database::postgres::org::Org;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    name: String,
    #[structopt(long, short)]
    force: bool,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(OrgRemoveRequest {
        name: action_option.name.into(),
        force: action_option.force.into(),
    });
    debug!("Request for org remove: {:?}", &request);

    let response = client.org_remove(request).await;
    match response {
        Err(_e) => {
            eprintln!("Org not found");
            Ok(())
        }
        Ok(o) => {
            let org_proto = o.into_inner();

            debug!("RESPONSE = {:?}", &org_proto);

            // By default, format the response into a table
            let mut table = Table::new();
            table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

            // Print the header row
            table.set_titles(row![bc=> "Org Name", "Active", "Created", "Last Update"]);

            let org = Org::from(org_proto.clone());
            table.add_row(row![
                org.name,
                &format!(
                    "{:?}",
                    orbital_database::postgres::schema::ActiveState::Deleted
                ),
                org.created,
                org.last_update
            ]);

            // Print the table to stdout
            table.printstd();
            Ok(())
        }
    }
}
