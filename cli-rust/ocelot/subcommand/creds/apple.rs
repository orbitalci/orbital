//ocelot creds apple add -acct my_kewl_acct -zip=/Users/jessishank/jessdev.developerprofile
//ocelot creds apple list

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    /// Add Apple Developer account (as Xcode exported .developerprofile)
    Add(AppleAddOption),
    /// Delete a Apple Developer account
    Delete,
    /// List registered Apple Developer account(s)
    List,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AppleAddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
    /// File path to Xcode exported Apple developer profile
    #[structopt(name = "developerprofile", short = "f", long = "file")]
    file_path: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AppleOptions {
    #[structopt(flatten)]
    action: ResourceAction,

    #[structopt(long)]
    acct: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler() {
    println!("Placeholder for handling Apple developer creds");
}