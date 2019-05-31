extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    /// Add a Version Control System
    Add(VcsAddOption),
    /// Delete a Version Control System
    Delete,
    /// List registered Version Control System
    List,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct VcsAddOption {
    /// File path to yaml configuration file
    #[structopt(name = "config file", short = "f", long = "file")]
    file_path: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct VcsOptions {
    #[structopt(flatten)]
    action: ResourceAction,

    #[structopt(long)]
    acct: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler() {
    println!("Placeholder for handling VCS creds");
}
