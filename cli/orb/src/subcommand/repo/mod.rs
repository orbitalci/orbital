use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::code::{client::CodeServiceClient, GitRepoAddRequest};

use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

pub mod add;
pub mod get;
pub mod list;
pub mod remove;
pub mod update;

/// Local options for customizing repo request
#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,

    #[structopt(subcommand)]
    pub action: Action,
}

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    Add(add::ActionOption),
    Get(get::ActionOption),
    Update(update::ActionOption),
    Remove(remove::ActionOption),
    List(list::ActionOption),
    // Enable,
    // Disable,
}

/// *Not yet implemented* Backend calls for managing repo resources
pub async fn subcommand_handler(
    global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    match local_option.clone().action {
        Action::Add(action_option) => {
            add::action_handler(global_option, local_option, action_option).await
        }
        Action::Get(action_option) => {
            get::action_handler(global_option, local_option, action_option).await
        }
        Action::Update(action_option) => {
            update::action_handler(global_option, local_option, action_option).await
        }
        Action::Remove(action_option) => {
            remove::action_handler(global_option, local_option, action_option).await
        }
        Action::List(action_option) => {
            list::action_handler(global_option, local_option, action_option).await
        }
    }

    //let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    //let request = Request::new(GitRepoAddRequest {
    //    org: "org_name_goes_here".into(),
    //    uri: "uri_goes_here".into(),
    //    ..Default::default()
    //});

    //let response = client.git_repo_add(request).await?;

    //println!("RESPONSE = {:?}", response);

    //Ok(())
}
