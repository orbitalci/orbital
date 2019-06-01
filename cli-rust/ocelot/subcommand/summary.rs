extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    // build-id will provide the same functionality that the `status` subcommand did.
    /// Retrieve status for specific build 
    #[structopt(name = "build id", long)]
    build_id: Option<u32>,
    /// Retrieve status for builds from the provided acct-repo
    #[structopt(long)]
    acct_repo: Option<String>,
    /// Retrieve status for builds from the provided branch.
    #[structopt(long)]
    branch : Option<String>,
    /// Retrieve status for builds with the provided commit hash
    #[structopt(long)]
    hash : Option<String>,
    /// Limit to last N runs
    #[structopt(long)]
    limit : Option<u32>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling summary");
}