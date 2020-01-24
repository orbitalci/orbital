use structopt::StructOpt;

use crate::{org::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::organization::{
    organization_service_client::OrganizationServiceClient, OrgUpdateRequest,
};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use orbital_database::postgres::schema::ActiveState;

use log::debug;

use prettytable::{cell, format, row, Table};

use orbital_database::postgres::org::Org;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    name: String,
    update_name: String,
    #[structopt(long, short, default_value = "enabled", possible_values = &ActiveState::variants())]
    active_state: ActiveState,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(OrgUpdateRequest {
        name: action_option.name.into(),
        update_name: action_option.update_name.into(),
        active_state: action_option.active_state.into(),
    });

    debug!("Request for org update: {:?}", &request);

    let response = client.org_update(request).await;
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
            table.set_titles(row!["Org Name", "Active", "Created", "Last Update"]);

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
