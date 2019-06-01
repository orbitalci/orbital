//  ocelot poll delete -acct-repo level11consulting/ocelog
//      -- I'm not sure we should force the --acct-repo flag
// Should have add as a subcommand instead of the current functionality

// ocelot poll list might want to filter by account

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,

    /// Use the provided acct-repo
    #[structopt(long)]
    acct_repo: Option<String>,

    /// Cron string
    #[structopt(long = "cron")]
    cron_string : Option<String>,

    /// Comma-separated list of branches
    #[structopt(alias = "branches")]
    branch: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct ListOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct DeleteOption {
    /// Delete the poll schedule for the provided account/repo
    #[structopt(long)]
    acct_repo: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    /// Add a polling schedule
    Add(AddOption),
    /// Delete a polling schedule
    #[structopt(alias = "rm")]
    Delete(DeleteOption),
    /// List the polling schedules
    #[structopt(alias = "ls")]
    List(ListOption),
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    #[structopt(flatten)]
    action: ResourceAction,

    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling polling schedules");
}
