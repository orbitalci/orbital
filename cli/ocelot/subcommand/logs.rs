//ocelot logs --hash <git_hash>
// add --build-id
// add --acct-repo

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Build ID
    #[structopt(name = "build id", long)]
    build_id: Option<u32>,
    /// Retrieve logs for account/repo. Otherwise try to auto-detect from current working directory
    #[structopt(long)]
    acct_repo: Option<String>,
    /// Retrieve logs for the provided branch. Without build-id or hash, will default to latest commit in branch
    #[structopt(long)]
    branch : Option<String>,
    /// Retrieve logs for the provided commit hash. Otherwise, default to latest build
    #[structopt(long)]
    hash : Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling logs");
}
