use structopt::StructOpt;

use crate::{org::SubcommandOption, GlobalOption};

use orbital_headers::organization::organization_service_client::OrganizationServiceClient;
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use prettytable::{cell, format, row, Table};

use color_eyre::eyre::Result;
use orbital_database::postgres::org::Org;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    _action_option: ActionOption,
) -> Result<()> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(());
    debug!("Request for org list: {:?}", &request);

    let response = client.org_list(request).await?.into_inner();

    // By default, format the response into a table
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

    // Print the header row
    table.set_titles(row![bc => "Org Name", "Active", "Created", "Last Update"]);

    match response.orgs.len() {
        0 => {
            println!("No Orgs found");
        }
        _ => {
            for org_proto in response.orgs {
                let org = Org::from(org_proto);
                table.add_row(row![
                    org.name,
                    &format!("{:?}", org.active_state),
                    org.created,
                    org.last_update
                ]);
            }
        }
    }

    // Print the table to stdout
    table.printstd();

    //println!("RESPONSE = {:?}", response);
    Ok(())
}
