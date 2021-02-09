use structopt::StructOpt;

use crate::{org::SubcommandOption, GlobalOption};

use orbital_headers::organization::{
    organization_service_client::OrganizationServiceClient, OrgAddRequest,
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
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(OrgAddRequest {
        name: action_option.name.into(),
    });
    debug!("Request for org add: {:?}", &request);

    let response = client.org_add(request).await;

    match response {
        Err(_e) => {
            eprintln!("Error adding Org");
            Ok(())
        }
        Ok(o) => {
            let org_proto = o.into_inner();

            debug!("RESPONSE = {:?}", &org_proto);

            // By default, format the response into a table
            let mut table = Table::new();
            table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

            // Print the header row
            table.set_titles(row![bc => "Org Name", "Active", "Created", "Last Update"]);

            let org = Org::from(org_proto.clone());
            table.add_row(row![
                org.name,
                &format!("{:?}", org.active_state),
                org.created,
                org.last_update
            ]);

            // Print the table to stdout
            table.printstd();
            Ok(())
        }
    }
}
