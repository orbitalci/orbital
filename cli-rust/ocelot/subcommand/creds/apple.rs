//ocelot creds apple add -acct my_kewl_acct -zip=/Users/jessishank/jessdev.developerprofile
//ocelot creds apple list

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
    /// File path to Xcode exported Apple developer profile
    #[structopt(name = "developerprofile", short = "f", long = "file")]
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
    /// Add Apple Developer account (as Xcode exported .developerprofile)
    Add(AddOption),
    /// Delete a Apple Developer account
    #[structopt(alias = "rm")]
    Delete,
    /// List registered Apple Developer account(s)
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
    println!("Placeholder for handling Apple developer creds");
}
