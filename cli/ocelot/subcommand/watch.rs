extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Unwatch the git repo
    #[structopt(long)]
    unwatch: Option<bool>,

    /// Watch the provided account/repo for new commits
    #[structopt(long)]
    acct_repo: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling watch");
}
