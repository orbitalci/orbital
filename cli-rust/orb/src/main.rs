//#[macro_use]
extern crate structopt;

extern crate clap;

use clap::arg_enum;
use structopt::StructOpt;

// This is for autocompletion
use structopt::clap::Shell;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "snake_case")]
pub enum Actions {
    Add,
    Delete,
    Update,
    Get,
    List,
    Enable,
    Disable,
}

arg_enum! {
    #[derive(Debug, StructOpt)]
    #[structopt(rename_all = "snake_case")]
    /// Supported types of secrets
    pub enum SecretType {
        DockerRegistry,
        Npm,
        Pypi,
        Maven,
        Ssh,
        Helm,
        Kubernetes,
        AppleDev,
        EnvVar,
        File,
    }
}

arg_enum! {
    #[derive(Debug, StructOpt)]
    #[structopt(rename_all = "snake_case")]
    /// Supported hosted git version control providers
    pub enum VcsType {
        Bitbucket,
        Github,
    }
}

arg_enum! {
    #[derive(Debug, StructOpt)]
    #[structopt(rename_all = "snake_case")]
    /// Supported notification methods
    pub enum NotifierType {
        Slack,
        Webhook,
    }
}

arg_enum! {
    #[derive(Debug, StructOpt)]
    #[structopt(rename_all = "snake_case")]
    /// Supported repo methods
    pub enum RepoType {
        Poll,
    }
}

///#[derive(Debug, StructOpt)]
///#[structopt(rename_all = "snake_case")]
////// User Resources
///pub enum UserResource {
///    Org,
///    Vcs(VcsType),
///    Notifier(NotifierType),
///    Secret(SecretType),
///    Repo(RepoType),
///}

//// TODO: Needs server stuff
//#[derive(Debug, StructOpt)]
//#[structopt(rename_all = "snake_case")]
///// Operator Resources
//pub enum OperatorResource {
//    /// Enable a resource
//    Enable(UserResource),
//    /// Disable a resource
//    Disable(UserResource),
//    Summary,
//    /// Trigger an event
//    Trigger,
//    /// Print the status of a particular resource
//    Status,
//}

// TODO: Need an easy to maintain way to map actions add/modify/destroy/view/enable/disable to resources
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "snake_case")]
pub enum Command {
    SecretType(Actions),
    ///// Add a resource
    //Add(UserResource),

    ///// Update a resource
    //Update(UserResource),

    ///// Delete a resource
    //Delete(UserResource),

    ///// Get a details of a resource
    //Get(UserResource),

    ///// Start a build job for registered repo
    //Build,

    ///// View the summary of a build
    //Summary,

    ///// Send SIGKILL (ctrl+c) signal to a running build to stop
    //Cancel,

    ///// Print the logs of a build
    //Logs,

    ///// Tools for OrbitalCI operators
    //Operator(OperatorResource),

    /// Create a stub orb.yml
    Init,

    ///// Print the version for orb
    //Version,


}

#[derive(Debug, StructOpt)]
#[structopt(name = "orb")]
/// The OrbitalCI command line interface
pub struct ApplicationArguments {
    #[structopt(subcommand)]
    pub command: Command,
    #[structopt(long = "consul-addr")]
    /// http api address of Consul. Specified as URI with scheme (e.g http://127.0.0.1:8500)
    pub consul_addr: Option<String>,
    #[structopt(long = "vault-addr")]
    /// http api address of Vault. Specified as URI with scheme (e.g http://127.0.0.1:8200)
    pub vault_addr: Option<String>,
    #[structopt(long = "vault-token")]
    pub vault_token: Option<String>,

    #[structopt(long = "nsqd-addr")]
    pub nsqd_addr: Option<String>,
    #[structopt(long = "nsqlookupd-addr")]
    pub nsqlookupd_addr: Option<String>,


    #[structopt(long = "debug")]
    pub debug : bool,

}

fn main() {
    // generate `bash` completions in "target" directory
    ApplicationArguments::clap().gen_completions(env!("CARGO_PKG_NAME"), Shell::Bash, "target");   

    let matches = ApplicationArguments::from_args();


    println!("{:?}", matches);
}