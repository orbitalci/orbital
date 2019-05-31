pub mod apple;
pub mod env;
pub mod helmrepo;
pub mod k8s;
pub mod notify;
pub mod repo;
pub mod ssh;
pub mod vcs;

use structopt::StructOpt;

// FIXME: Delete this after we move this into each of the type handlers
#[derive(Debug, StructOpt, Copy, Clone)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    Add,
    Delete,
    List,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum CredType {
    Apple(apple::AppleOptions),
    Env(ResourceAction),
    Helmrepo(ResourceAction),
    K8s(ResourceAction),
    Notify(ResourceAction),
    Repo(ResourceAction),
    Ssh(ResourceAction),
    Vcs(vcs::VcsOptions),
}

// Handle the command line control flow
pub fn subcommand_handler() {
    println!("Placeholder for handling creds");
}