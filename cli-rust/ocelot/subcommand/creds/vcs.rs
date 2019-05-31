extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// File path to yaml configuration file
    #[structopt(name = "config file", short = "f", long = "file")]
    file_path: Option<String>,
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
pub enum ResourceAction {
    /// Add a Version Control System
    Add(AddOption),
    /// Delete a Version Control System
    #[structopt(alias = "rm")]
    Delete,
    /// List registered Version Control System
    #[structopt(alias = "ls")]
    List(ListOption),
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    #[structopt(flatten)]
    action: ResourceAction,

    #[structopt(long)]
    acct: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling VCS creds");
}