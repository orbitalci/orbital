//ocelot creds notify add --identifier L11_SLACK --acctname level11consulting --url https://hooks.slack.com/services/T0DFsdSBA/345PPRP9C/5hUe12345v6BrxfSJt --detail-url https://ocelot.mysite.io

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
    /// Kubernetes cluster name
    #[structopt(name = "Slack org name", long)]
    slack_name: Option<String>,
    /// File path to yaml containing env vars
    #[structopt(name = "Kubernetes config (yaml)", short = "f", long = "file")]
    webhook_url: Option<String>,
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
    /// 
    Add(AddOption),
    /// 
    #[structopt(alias = "rm")]
    Delete,
    /// 
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
    println!("Placeholder for handling Slack creds");
}
