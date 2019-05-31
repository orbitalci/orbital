//ocelot creds helmrepo add -acct my_kewl_acct -repo-name shankj3_charts -helm-url https://github.io/shankj3_helm_repository
//ocelot creds helmrepo list -account <ACCT_NAME>

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
    /// Helm repo name (logical)
    #[structopt(name = "Helm repo name", long)]
    helm_name: Option<String>,
    /// Helm repo url
    #[structopt(name = "Helm repo url", long)]
    helm_url: Option<String>,
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
    println!("Placeholder for handling Helm repo creds");
}
